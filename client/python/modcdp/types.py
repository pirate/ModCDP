from __future__ import annotations

from collections.abc import Callable, Mapping
from queue import Queue
from typing import Any, Literal, Protocol, TypeAlias, TypedDict

from typing_extensions import NotRequired

JsonPrimitive: TypeAlias = None | bool | int | float | str
JsonValue: TypeAlias = JsonPrimitive | list["JsonValue"] | dict[str, "JsonValue"]
JsonObject: TypeAlias = dict[str, JsonValue]

ProtocolParams: TypeAlias = Mapping[str, JsonValue]
ProtocolResult: TypeAlias = dict[str, JsonValue]
ProtocolPayload: TypeAlias = dict[str, JsonValue]
MessageParams: TypeAlias = Mapping[str, object]
ModCDPRoutes: TypeAlias = dict[str, str]


class _ModCDPAddCustomCommandRequired(TypedDict):
    name: str


class ModCDPAddCustomCommandParams(_ModCDPAddCustomCommandRequired, total=False):
    expression: str | None
    params_schema: JsonValue
    result_schema: JsonValue


class _ModCDPAddCustomEventObjectRequired(TypedDict):
    name: str


class ModCDPAddCustomEventObjectParams(_ModCDPAddCustomEventObjectRequired, total=False):
    event_schema: JsonValue


ModCDPAddCustomEventParams: TypeAlias = str | ModCDPAddCustomEventObjectParams


class _ModCDPAddMiddlewareRequired(TypedDict):
    phase: Literal["request", "response", "event"]
    expression: str


class ModCDPAddMiddlewareParams(_ModCDPAddMiddlewareRequired, total=False):
    name: str


class ModCDPPingLatency(TypedDict):
    sent_at: int
    received_at: int | float | None
    returned_at: int
    round_trip_ms: int
    service_worker_ms: int | float | None
    return_path_ms: int | float | None


class ModCDPConnectTiming(TypedDict):
    started_at: int
    upstream_mode: str | None
    upstream_endpoint_kind: Literal["raw_cdp", "modcdp_server"]
    transport_started_at: int
    transport_connected_at: int
    transport_duration_ms: int
    extension_source: NotRequired[str | None]
    extension_started_at: NotRequired[int]
    extension_completed_at: NotRequired[int]
    extension_duration_ms: NotRequired[int]
    connected_at: int
    duration_ms: int


class ModCDPCommandTiming(TypedDict):
    method: str
    target: str
    started_at: int
    completed_at: int
    duration_ms: int


class ModCDPRawTiming(TypedDict):
    method: str
    started_at: int
    completed_at: int
    duration_ms: int


class ModCDPServerConfig(TypedDict, total=False):
    loopback_cdp_url: str | None
    routes: ModCDPRoutes
    cdp_send_timeout_ms: int
    loopback_execution_context_timeout_ms: int
    ws_connect_error_settle_timeout_ms: int
    browser_token: str | None
    custom_commands: list[ModCDPAddCustomCommandParams]
    custom_events: list[ModCDPAddCustomEventObjectParams]
    custom_middlewares: list[ModCDPAddMiddlewareParams]


RuntimeEvaluateParams: TypeAlias = dict[str, JsonValue]


class _TranslatedStepRequired(TypedDict):
    method: str


class TranslatedStep(_TranslatedStepRequired, total=False):
    params: MessageParams
    sessionId: str | None
    unwrap: Literal["evaluate"]


class TranslatedCommand(TypedDict):
    route: str
    target: Literal["direct_cdp", "service_worker"]
    steps: list[TranslatedStep]


class CdpError(TypedDict, total=False):
    message: str


class CdpMessage(TypedDict, total=False):
    id: int
    method: str
    params: MessageParams
    sessionId: str
    result: ProtocolResult
    error: CdpError


class TargetInfo(TypedDict):
    targetId: str
    type: str
    url: str


class ExtensionProbe(TypedDict):
    extension_id: str
    target_id: str
    url: str
    session_id: str


class ExtensionInfo(ExtensionProbe):
    source: str


class BorrowedExtensionInfo(ExtensionInfo, total=False):
    has_tabs: bool
    has_debugger: bool


class UnwrappedModCDPEvent(TypedDict):
    event: str
    data: ProtocolPayload
    sessionId: str | None


Handler: TypeAlias = Callable[[Any], Any]
PendingEntry: TypeAlias = tuple[str, Queue[CdpMessage]]


class WebSocketLike(Protocol):
    def send(self, payload: str) -> object: ...

    def recv(self) -> str | bytes | None: ...

    def close(self) -> object: ...
