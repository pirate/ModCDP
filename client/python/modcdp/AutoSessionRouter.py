from __future__ import annotations

import threading
from collections.abc import Callable, Mapping
from typing import Any


SendCDP = Callable[[str, dict[str, Any], str | None], dict[str, Any]]


class AutoSessionRouter:
    def __init__(self, send: SendCDP, defaultExecutionContextTimeoutMs: Callable[[], int]) -> None:
        self.send = send
        self.defaultExecutionContextTimeoutMs = defaultExecutionContextTimeoutMs
        self.target_sessions: dict[str, str] = {}
        self.session_targets: dict[str, dict[str, Any]] = {}
        self.execution_contexts: dict[str, int] = {}
        self._execution_context_waiters: dict[str, list[tuple[threading.Event, dict[str, int]]]] = {}
        self._lock = threading.RLock()

    def sessionIdForTarget(self, target_id: str) -> str | None:
        with self._lock:
            return self.target_sessions.get(target_id)

    def attachToTarget(self, target_id: str) -> str | None:
        existing_session_id = self.sessionIdForTarget(target_id)
        if existing_session_id is not None:
            return existing_session_id
        result = self.send("Target.attachToTarget", {"targetId": target_id, "flatten": True}, None)
        session_id = result.get("sessionId")
        return session_id if isinstance(session_id, str) and session_id else None

    def recordProtocolEvent(self, method: str, data: object, session_id: str | None) -> None:
        event_data = dict(data) if isinstance(data, Mapping) else {}
        if method == "Target.attachedToTarget":
            attached_session_id = event_data.get("sessionId") if isinstance(event_data.get("sessionId"), str) else session_id
            raw_target_info = event_data.get("targetInfo")
            target_info = dict(raw_target_info) if isinstance(raw_target_info, Mapping) else None
            target_id = target_info.get("targetId") if target_info else None
            if isinstance(attached_session_id, str) and isinstance(target_id, str) and target_info:
                with self._lock:
                    self.target_sessions[target_id] = attached_session_id
                    self.session_targets[attached_session_id] = target_info
        elif method == "Runtime.executionContextCreated":
            raw_context = event_data.get("context")
            context = raw_context if isinstance(raw_context, Mapping) else None
            context_id = context.get("id") if context else None
            if session_id and isinstance(context_id, int):
                self._recordExecutionContext(session_id, context_id)
        elif method == "Target.detachedFromTarget":
            detached_session_id = event_data.get("sessionId") if isinstance(event_data.get("sessionId"), str) else session_id
            if isinstance(detached_session_id, str):
                self._forgetSession(detached_session_id)

    def waitForExecutionContext(self, session_id: str | None, timeout_ms: int | None = None) -> int:
        effective_timeout_ms = timeout_ms if timeout_ms is not None else self.defaultExecutionContextTimeoutMs()
        if not session_id:
            raise RuntimeError("Cannot wait for a Runtime execution context without a session.")
        with self._lock:
            existing = self.execution_contexts.get(session_id)
            if existing is not None:
                return existing
            event = threading.Event()
            result: dict[str, int] = {}
            self._execution_context_waiters.setdefault(session_id, []).append((event, result))
        if not event.wait(effective_timeout_ms / 1000):
            with self._lock:
                waiters = self._execution_context_waiters.get(session_id, [])
                self._execution_context_waiters[session_id] = [item for item in waiters if item[0] is not event]
                if not self._execution_context_waiters[session_id]:
                    self._execution_context_waiters.pop(session_id, None)
            raise RuntimeError(f"Timed out waiting for Runtime.executionContextCreated for session {session_id}.")
        return result["context_id"]

    def _recordExecutionContext(self, session_id: str, context_id: int) -> None:
        with self._lock:
            self.execution_contexts[session_id] = context_id
            waiters = self._execution_context_waiters.pop(session_id, [])
        for event, result in waiters:
            result["context_id"] = context_id
            event.set()

    def _forgetSession(self, session_id: str) -> None:
        with self._lock:
            target_info = self.session_targets.pop(session_id, None)
            target_id = target_info.get("targetId") if target_info else None
            if isinstance(target_id, str):
                self.target_sessions.pop(target_id, None)
            self.execution_contexts.pop(session_id, None)
