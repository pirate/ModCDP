from __future__ import annotations

import json
import threading
import unittest
from queue import Queue
from typing import Any

from websocket import create_connection

from modcdp.router.AutoSessionRouter import AutoSessionRouter
from modcdp.launcher.LocalBrowserLauncher import LocalBrowserLauncher


class AutoSessionRouterTests(unittest.TestCase):
    def test_rejects_pending_execution_context_waiters_when_session_detaches(self) -> None:
        router = AutoSessionRouter(lambda _method, _params, _session_id: {}, lambda: 5_000)
        result: Queue[int | BaseException] = Queue()
        threading.Thread(
            target=lambda: _put_result(result, lambda: router.waitForExecutionContext("detached-session", 5_000)),
            daemon=True,
        ).start()

        router.recordProtocolEvent(
            "Target.attachedToTarget",
            {"sessionId": "detached-session", "targetInfo": {"targetId": "target-1", "type": "page"}},
            None,
        )
        router.recordProtocolEvent("Target.detachedFromTarget", {"sessionId": "detached-session"}, None)
        router.recordProtocolEvent(
            "Runtime.executionContextCreated",
            {"context": {"id": 42}},
            "detached-session",
        )

        error = result.get(timeout=1)
        self.assertIsInstance(error, RuntimeError)
        self.assertIn("Runtime execution context wait cancelled because session detached-session detached.", str(error))
        self.assertIsNone(router.sessionIdForTarget("target-1"))
        self.assertNotIn("detached-session", router.execution_contexts)

    def test_bounds_detached_session_guards_and_clears_them_when_session_reattaches(self) -> None:
        router = AutoSessionRouter(lambda _method, _params, _session_id: {}, lambda: 5_000)

        for index in range(1034):
            router.recordProtocolEvent("Target.detachedFromTarget", {"sessionId": f"detached-session-{index}"}, None)

        self.assertLessEqual(len(router._detached_sessions), 1024)

        recent_session_id = "detached-session-1033"
        router.recordProtocolEvent("Runtime.executionContextCreated", {"context": {"id": 42}}, recent_session_id)
        self.assertNotIn(recent_session_id, router.execution_contexts)

        router.recordProtocolEvent(
            "Target.attachedToTarget",
            {"sessionId": recent_session_id, "targetInfo": {"targetId": "target-reattached", "type": "page"}},
            None,
        )
        router.recordProtocolEvent("Runtime.executionContextCreated", {"context": {"id": 43}}, recent_session_id)

        self.assertEqual(router.sessionIdForTarget("target-reattached"), recent_session_id)
        self.assertEqual(router.execution_contexts[recent_session_id], 43)

    def test_tracks_real_target_sessions_and_execution_contexts(self) -> None:
        chrome = LocalBrowserLauncher({"headless": True}).launch()
        ws = create_connection(str(chrome["cdp_url"]), timeout=10)
        lock = threading.Lock()
        next_id = 0
        pending: dict[int, Queue[dict[str, Any]]] = {}
        closed = False

        def send(method: str, params: dict[str, Any] | None = None, session_id: str | None = None) -> dict[str, Any]:
            nonlocal next_id
            with lock:
                next_id += 1
                msg_id = next_id
                done: Queue[dict[str, Any]] = Queue()
                pending[msg_id] = done
            message = {"id": msg_id, "method": method, "params": params or {}}
            if session_id:
                message["sessionId"] = session_id
            ws.send(json.dumps(message))
            response = done.get(timeout=10)
            if response.get("error"):
                raise RuntimeError(json.dumps(response["error"]))
            result = response.get("result")
            return result if isinstance(result, dict) else {}

        router = AutoSessionRouter(send, lambda: 30_000)

        def reader() -> None:
            while not closed:
                try:
                    raw = ws.recv()
                except Exception:
                    return
                if not raw:
                    return
                try:
                    message = json.loads(raw)
                except json.JSONDecodeError:
                    return
                if isinstance(message.get("id"), int):
                    done = pending.pop(message["id"], None)
                    if done is not None:
                        done.put(message)
                    continue
                method = message.get("method")
                if isinstance(method, str):
                    session_id = message.get("sessionId")
                    router.recordProtocolEvent(method, message.get("params"), session_id if isinstance(session_id, str) else None)

        thread = threading.Thread(target=reader, daemon=True)
        thread.start()
        target_id: str | None = None
        try:
            send("Target.setAutoAttach", {"autoAttach": True, "waitForDebuggerOnStart": False, "flatten": True})
            send("Target.setDiscoverTargets", {"discover": True})
            created = send("Target.createTarget", {"url": "about:blank#modcdp-auto-session-router"})
            target_id = str(created["targetId"])
            session_id = _wait_for(lambda: router.sessionIdForTarget(target_id))
            context_result: Queue[int | BaseException] = Queue()
            threading.Thread(
                target=lambda: _put_result(context_result, lambda: router.waitForExecutionContext(session_id, 30_000)),
                daemon=True,
            ).start()
            send("Runtime.enable", {}, session_id)
            context_id = context_result.get(timeout=35)
            if isinstance(context_id, BaseException):
                raise context_id
            self.assertIsInstance(context_id, int)
            self.assertEqual(router.execution_contexts[session_id], context_id)

            send("Target.detachFromTarget", {"sessionId": session_id})
            _wait_for(lambda: None if router.sessionIdForTarget(target_id) else "detached")
        finally:
            if target_id:
                try:
                    send("Target.closeTarget", {"targetId": target_id})
                except Exception:
                    pass
            closed = True
            ws.close()
            chrome["close"]()


def _put_result(queue: Queue[int | BaseException], fn) -> None:
    try:
        queue.put(fn())
    except BaseException as error:
        queue.put(error)


def _wait_for(fn, timeout_s: float = 5) -> str:
    deadline = threading.Event()
    timer = threading.Timer(timeout_s, deadline.set)
    timer.start()
    try:
        while not deadline.is_set():
            value = fn()
            if value:
                return value
            deadline.wait(0.05)
    finally:
        timer.cancel()
    raise TimeoutError("timed out waiting for condition")


if __name__ == "__main__":
    unittest.main()
