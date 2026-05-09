/// <reference types="chrome" />

import { z } from "zod";

const isZodType = (value: unknown): value is z.ZodType =>
  value != null && typeof value === "object" && typeof (value as z.ZodType).parse === "function";

export const CdpCommandParamsSchema = z.object({}).passthrough();
export type CdpCommandParams = z.infer<typeof CdpCommandParamsSchema>;

export const CdpCommandResultSchema = z.object({}).passthrough();
export type CdpCommandResult = z.infer<typeof CdpCommandResultSchema>;

export const CdpEventParamsSchema = z.object({}).passthrough();
export type CdpEventParams = z.infer<typeof CdpEventParamsSchema>;

export const RuntimeBindingCalledEventSchema = z
  .object({
    name: z.string(),
    payload: z.string(),
    executionContextId: z.number().optional(),
  })
  .passthrough();
export type RuntimeBindingCalledEvent = z.infer<typeof RuntimeBindingCalledEventSchema>;

export const TargetAttachedToTargetEventSchema = z
  .object({
    sessionId: z.string(),
    targetInfo: z.object({ targetId: z.string() }).passthrough(),
    waitingForDebugger: z.boolean(),
  })
  .passthrough();
export type TargetAttachedToTargetEvent = z.infer<typeof TargetAttachedToTargetEventSchema>;

export const ModCDPRoutesSchema = z.object({}).catchall(z.string());
export type ModCDPRoutes = z.infer<typeof ModCDPRoutesSchema>;

export const ModCDPCustomPayloadSchema = z.object({}).passthrough();
export type ModCDPCustomPayload = z.infer<typeof ModCDPCustomPayloadSchema>;

export type ModCDPNamedValue = {
  cdp_command_name?: string;
  cdp_event_name?: string;
  id?: string;
  name?: string;
  meta?: () => { cdp_command_name?: unknown; cdp_event_name?: unknown; id?: unknown; name?: unknown } | undefined;
};

export function normalizeModCDPName(value: ModCDPName) {
  if (typeof value === "string") return value;
  const meta = typeof value?.meta === "function" ? value.meta() : undefined;
  const name =
    value?.cdp_command_name ??
    value?.cdp_event_name ??
    (typeof meta?.cdp_command_name === "string" ? meta.cdp_command_name : undefined) ??
    (typeof meta?.cdp_event_name === "string" ? meta.cdp_event_name : undefined) ??
    value?.id ??
    (typeof meta?.id === "string" ? meta.id : undefined) ??
    (typeof meta?.name === "string" ? meta.name : undefined) ??
    value?.name;
  if (typeof name !== "string" || !name) throw new Error("Expected a CDP name string or named CDP schema.");
  return name;
}

export const ModCDPNameSchema = z.custom<string | ModCDPNamedValue>((value) => {
  try {
    normalizeModCDPName(value as ModCDPName);
    return true;
  } catch {
    return false;
  }
});
export type ModCDPName = z.infer<typeof ModCDPNameSchema>;

export const ModCDPZodTypeSchema = z.custom<z.ZodType>(isZodType);
export type ModCDPZodType = z.infer<typeof ModCDPZodTypeSchema>;

export const ModCDPPayloadJsonSchemaSchema = z.record(z.string(), z.unknown());
export const ModCDPPayloadShapeSchema = z.record(z.string(), ModCDPZodTypeSchema);
export type ModCDPPayloadShape = z.infer<typeof ModCDPPayloadShapeSchema>;

export const ModCDPPayloadSchemaSpecSchema = z.union([
  ModCDPZodTypeSchema,
  ModCDPPayloadShapeSchema,
  ModCDPPayloadJsonSchemaSchema,
]);
export type ModCDPPayloadSchemaSpec = z.infer<typeof ModCDPPayloadSchemaSpecSchema>;

export function normalizeModCDPPayloadSchema(schema: ModCDPPayloadSchemaSpec | null | undefined) {
  if (!schema) return null;
  if (isZodType(schema)) return schema;
  if (Object.values(schema).every(isZodType)) return z.object(schema as ModCDPPayloadShape).passthrough();
  if (schema.type === "object") return z.object({}).passthrough();
  throw new Error("Unsupported payload schema; pass a Zod schema, Zod shape, or object JSON schema.");
}

export const ModCDPEvaluateParamsSchema = z.object({
  expression: z.string(),
  params: ModCDPCustomPayloadSchema.optional(),
  cdpSessionId: z.string().nullable().optional(),
});
export type ModCDPEvaluateParams = z.infer<typeof ModCDPEvaluateParamsSchema>;

export const ModCDPAddCustomCommandParamsSchema = z.object({
  name: ModCDPNameSchema,
  expression: z.string().nullable().optional(),
  params_schema: ModCDPPayloadSchemaSpecSchema.nullable().optional(),
  result_schema: ModCDPPayloadSchemaSpecSchema.nullable().optional(),
});
export type ModCDPAddCustomCommandParams = z.infer<typeof ModCDPAddCustomCommandParamsSchema>;

export const ModCDPAddCustomEventObjectParamsSchema = z.object({
  name: ModCDPNameSchema,
  event_schema: ModCDPPayloadSchemaSpecSchema.nullable().optional(),
});
export type ModCDPAddCustomEventObjectParams = z.infer<typeof ModCDPAddCustomEventObjectParamsSchema>;
export const ModCDPAddCustomEventParamsSchema = z.union([ModCDPZodTypeSchema, ModCDPAddCustomEventObjectParamsSchema]);
export type ModCDPAddCustomEventParams = z.infer<typeof ModCDPAddCustomEventParamsSchema>;

export const ModCDPAddMiddlewareParamsSchema = z.object({
  name: ModCDPNameSchema.optional(),
  phase: z.enum(["request", "response", "event"]),
  expression: z.string(),
});
export type ModCDPAddMiddlewareParams = z.infer<typeof ModCDPAddMiddlewareParamsSchema>;

export const ModCDPLaunchOptionsSchema = z.object({}).passthrough();
export type ModCDPLaunchOptions = z.infer<typeof ModCDPLaunchOptionsSchema>;

export const ModCDPUpstreamOptionsSchema = z
  .object({
    mode: z.enum(["ws", "pipe", "nativemessaging", "reversews", "nats"]).optional(),
    nats_url: z.string().nullable().optional(),
    nats_subject_prefix: z.string().nullable().optional(),
  })
  .passthrough();
export type ModCDPUpstreamOptions = z.infer<typeof ModCDPUpstreamOptionsSchema>;

export const ModCDPClientOptionsSchema = z
  .object({
    routes: ModCDPRoutesSchema.optional(),
  })
  .passthrough();
export type ModCDPClientOptions = z.infer<typeof ModCDPClientOptionsSchema>;

export const ModCDPServerOptionsSchema = z
  .object({
    loopback_cdp_url: z.string().nullable().optional(),
    routes: ModCDPRoutesSchema.optional(),
    browser_token: z.string().nullable().optional(),
    cdp_send_timeout_ms: z.number().positive().optional(),
    loopback_execution_context_timeout_ms: z.number().positive().optional(),
    ws_connect_error_settle_timeout_ms: z.number().positive().optional(),
  })
  .passthrough();
export type ModCDPServerOptions = z.infer<typeof ModCDPServerOptionsSchema>;

export const ModCDPConfigureParamsSchema = z.object({
  launch: ModCDPLaunchOptionsSchema.optional(),
  upstream: ModCDPUpstreamOptionsSchema.optional(),
  client: ModCDPClientOptionsSchema.optional(),
  server: ModCDPServerOptionsSchema.optional(),
  custom_commands: z.array(ModCDPAddCustomCommandParamsSchema).optional(),
  custom_events: z.array(ModCDPAddCustomEventObjectParamsSchema).optional(),
  custom_middlewares: z.array(ModCDPAddMiddlewareParamsSchema).optional(),
});
export type ModCDPConfigureParams = z.infer<typeof ModCDPConfigureParamsSchema>;

export const ModCDPPingParamsSchema = z.object({
  sent_at: z.number().optional(),
});
export type ModCDPPingParams = z.infer<typeof ModCDPPingParamsSchema>;

export const ModCDPPongEventSchema = z.object({
  sent_at: z.number(),
  received_at: z.number(),
  from: z.string(),
});
export type ModCDPPongEvent = z.infer<typeof ModCDPPongEventSchema>;

export const ModCDPPingLatencySchema = z.object({
  sent_at: z.number(),
  received_at: z.number().nullable(),
  returned_at: z.number(),
  round_trip_ms: z.number(),
  service_worker_ms: z.number().nullable(),
  return_path_ms: z.number().nullable(),
});
export type ModCDPPingLatency = z.infer<typeof ModCDPPingLatencySchema>;

export const ModCDPCommandParamsSchema = z.union([
  ModCDPEvaluateParamsSchema,
  ModCDPAddCustomCommandParamsSchema,
  ModCDPAddCustomEventParamsSchema,
  ModCDPAddMiddlewareParamsSchema,
  ModCDPConfigureParamsSchema,
  ModCDPPingParamsSchema,
  ModCDPCustomPayloadSchema,
]);
export type ModCDPCommandParams = z.infer<typeof ModCDPCommandParamsSchema>;

export const ModCDPCommandResultSchema = z.union([
  z.object({ ok: z.boolean() }).passthrough(),
  ModCDPCustomPayloadSchema,
]);
export type ModCDPCommandResult = z.infer<typeof ModCDPCommandResultSchema>;

export const ModCDPEvaluateResponseSchema = z.unknown();
export type ModCDPEvaluateResponse = z.infer<typeof ModCDPEvaluateResponseSchema>;

export const ModCDPAddCustomCommandResponseSchema = z
  .object({
    name: z.string(),
    registered: z.boolean(),
  })
  .passthrough();
export type ModCDPAddCustomCommandResponse = z.infer<typeof ModCDPAddCustomCommandResponseSchema>;

export const ModCDPAddCustomEventResponseSchema = z
  .object({
    name: z.string(),
    registered: z.boolean(),
  })
  .passthrough();
export type ModCDPAddCustomEventResponse = z.infer<typeof ModCDPAddCustomEventResponseSchema>;

export const ModCDPAddMiddlewareResponseSchema = z
  .object({
    name: z.string(),
    phase: z.enum(["request", "response", "event"]),
    registered: z.boolean(),
  })
  .passthrough();
export type ModCDPAddMiddlewareResponse = z.infer<typeof ModCDPAddMiddlewareResponseSchema>;

export const ModCDPConfigureResponseSchema = z
  .object({
    loopback_cdp_url: z.string().nullable().optional(),
    routes: ModCDPRoutesSchema,
  })
  .passthrough();
export type ModCDPConfigureResponse = z.infer<typeof ModCDPConfigureResponseSchema>;

export const ModCDPPingResponseSchema = z
  .object({
    ok: z.boolean(),
  })
  .passthrough();
export type ModCDPPingResponse = z.infer<typeof ModCDPPingResponseSchema>;

export const ModCDPBindingPayloadSchema = z.object({
  event: z.string(),
  data: z.unknown(),
  cdpSessionId: z.string().nullable().optional(),
});
export type ModCDPBindingPayload = z.infer<typeof ModCDPBindingPayloadSchema>;

export const CdpDebuggeeCommandParamsSchema = ModCDPCustomPayloadSchema.extend({
  debuggee: z.custom<chrome.debugger.Debuggee>().nullable().optional(),
  tabId: z.number().nullable().optional(),
  targetId: z.string().nullable().optional(),
  extensionId: z.string().nullable().optional(),
});
export type CdpDebuggeeCommandParams = z.infer<typeof CdpDebuggeeCommandParamsSchema>;

export const ProtocolParamsSchema = z.union([CdpCommandParamsSchema, ModCDPCommandParamsSchema]);
export type ProtocolParams = z.infer<typeof ProtocolParamsSchema>;

export const ProtocolResultSchema = z.union([CdpCommandResultSchema, ModCDPCommandResultSchema]);
export type ProtocolResult = z.infer<typeof ProtocolResultSchema>;

export const ProtocolEventParamsSchema = z.union([
  CdpEventParamsSchema,
  ModCDPPongEventSchema,
  ModCDPCustomPayloadSchema,
]);
export type ProtocolEventParams = z.infer<typeof ProtocolEventParamsSchema>;

export const ProtocolPayloadSchema = z.union([
  ProtocolParamsSchema,
  ProtocolResultSchema,
  ProtocolEventParamsSchema,
  ModCDPBindingPayloadSchema,
  z.null(),
]);
export type ProtocolPayload = z.infer<typeof ProtocolPayloadSchema>;

export const ModCDPCustomCommandRegistrationSchema = ModCDPAddCustomCommandParamsSchema.extend({
  expression: z.string().nullable().optional(),
  handler:
    z.custom<
      (params: ProtocolParams, cdpSessionId: string | null, method?: string) => ProtocolResult | Promise<ProtocolResult>
    >(),
});
export type ModCDPCustomCommandRegistration = z.infer<typeof ModCDPCustomCommandRegistrationSchema>;

export const ModCDPCustomEventRegistrationSchema = ModCDPAddCustomEventObjectParamsSchema;
export type ModCDPCustomEventRegistration = z.infer<typeof ModCDPCustomEventRegistrationSchema>;

export const ModCDPMiddlewareRegistrationSchema = ModCDPAddMiddlewareParamsSchema.extend({
  expression: z.string().nullable().optional(),
  handler:
    z.custom<
      (
        payload: ProtocolPayload,
        next: (payload?: ProtocolPayload) => Promise<ProtocolPayload>,
        context: ModCDPCustomPayload,
      ) => ProtocolPayload | Promise<ProtocolPayload>
    >(),
});
export type ModCDPMiddlewareRegistration = z.infer<typeof ModCDPMiddlewareRegistrationSchema>;

export const CdpErrorSchema = z
  .object({
    code: z.number().optional(),
    message: z.string(),
    data: z.unknown().optional(),
  })
  .passthrough();
export type CdpError = z.infer<typeof CdpErrorSchema>;

export const CdpCommandMessageSchema = z
  .object({
    id: z.number(),
    method: z.string(),
    params: ProtocolParamsSchema.optional(),
    sessionId: z.string().optional(),
  })
  .passthrough();
export type CdpCommandMessage = z.infer<typeof CdpCommandMessageSchema>;

export const CdpResponseMessageSchema = z
  .object({
    id: z.number(),
    result: ProtocolResultSchema.optional(),
    error: CdpErrorSchema.optional(),
    sessionId: z.string().optional(),
  })
  .passthrough();
export type CdpResponseMessage = z.infer<typeof CdpResponseMessageSchema>;

export const CdpEventMessageSchema = z
  .object({
    method: z.string(),
    params: ProtocolEventParamsSchema.optional(),
    sessionId: z.string().optional(),
  })
  .passthrough();
export type CdpEventMessage = z.infer<typeof CdpEventMessageSchema>;

export const CdpMessageSchema = z.union([CdpCommandMessageSchema, CdpResponseMessageSchema, CdpEventMessageSchema]);
export type CdpMessage = z.infer<typeof CdpMessageSchema>;

export const TranslatedStepSchema = z
  .object({
    method: z.string(),
    params: ProtocolParamsSchema.optional(),
    sessionId: z.string().nullable().optional(),
    unwrap: z.literal("runtime").optional(),
  })
  .passthrough();
export type TranslatedStep = z.infer<typeof TranslatedStepSchema>;

export const TranslatedCommandSchema = z
  .object({
    route: z.string(),
    target: z.enum(["direct_cdp", "service_worker"]),
    steps: z.array(TranslatedStepSchema),
  })
  .passthrough();
export type TranslatedCommand = z.infer<typeof TranslatedCommandSchema>;

export const UnwrappedModCDPEventSchema = z
  .object({
    event: z.string(),
    data: ProtocolPayloadSchema,
    sessionId: z.string().nullable(),
  })
  .passthrough();
export type UnwrappedModCDPEvent = z.infer<typeof UnwrappedModCDPEventSchema>;

export const ProxyPendingSchema = z
  .object({
    kind: z.string(),
    client_id: z.number().optional(),
    client_session_id: z.string().nullable().optional(),
    event_name: z.string().optional(),
    resolve: z.custom<(value: ProtocolResult) => void>().optional(),
    reject: z.custom<(error: Error) => void>().optional(),
  })
  .passthrough();
export type ProxyPending = z.infer<typeof ProxyPendingSchema>;

export const ProxyUpstreamStateSchema = z
  .object({
    url: z.string(),
    launched: z
      .custom<Awaited<ReturnType<import("../bridge/BrowserLauncher.js").BrowserLauncher["launch"]>>>()
      .nullable(),
    launch_promise: z
      .promise(z.custom<Awaited<ReturnType<import("../bridge/BrowserLauncher.js").BrowserLauncher["launch"]>>>())
      .nullable()
      .optional(),
  })
  .passthrough();
export type ProxyUpstreamState = z.infer<typeof ProxyUpstreamStateSchema>;

export type ProxyRawData = Buffer | ArrayBuffer | Buffer[];
export type ProxyWebSocketLike = {
  CLOSED: number;
  CLOSING: number;
  readyState: number;
  close(code?: number, reason?: string | Buffer): void;
  send(data: string): void;
};

export const ProxyConnectionStateSchema = z.object({
  client: z.custom<ProxyWebSocketLike>(),
  upstream: z.custom<ProxyWebSocketLike>(),
  next_upstream_id: z.number(),
  pending: z.custom<Map<number, ProxyPending>>(),
  ext_session_id: z.string().nullable(),
  ext_target_id: z.string().nullable(),
  ext_execution_context_id: z.number().nullable(),
  hidden_session_ids: z.custom<Set<string>>(),
  hidden_target_ids: z.custom<Set<string>>(),
  target_session_ids: z.custom<Map<string, string>>(),
  client_session_ids: z.custom<Set<string>>(),
  forward_mirrored_upstream_events: z.boolean(),
  bootstrapped: z.boolean(),
  closing: z.boolean(),
  queued_from_client: z.array(z.custom<ProxyRawData>()),
});
export type ProxyConnectionState = z.infer<typeof ProxyConnectionStateSchema>;

export const Mod = {
  Routes: ModCDPRoutesSchema,
  CustomPayload: ModCDPCustomPayloadSchema,
  Name: ModCDPNameSchema,
  ZodType: ModCDPZodTypeSchema,
  PayloadShape: ModCDPPayloadShapeSchema,
  PayloadSchemaSpec: ModCDPPayloadSchemaSpecSchema,
  EvaluateParams: ModCDPEvaluateParamsSchema,
  AddCustomCommandParams: ModCDPAddCustomCommandParamsSchema,
  AddCustomEventObjectParams: ModCDPAddCustomEventObjectParamsSchema,
  AddCustomEventParams: ModCDPAddCustomEventParamsSchema,
  AddMiddlewareParams: ModCDPAddMiddlewareParamsSchema,
  LaunchOptions: ModCDPLaunchOptionsSchema,
  UpstreamOptions: ModCDPUpstreamOptionsSchema,
  ClientOptions: ModCDPClientOptionsSchema,
  ServerOptions: ModCDPServerOptionsSchema,
  ConfigureParams: ModCDPConfigureParamsSchema,
  PingParams: ModCDPPingParamsSchema,
  PongEvent: ModCDPPongEventSchema,
  PingLatency: ModCDPPingLatencySchema,
  CommandParams: ModCDPCommandParamsSchema,
  CommandResult: ModCDPCommandResultSchema,
  EvaluateResponse: ModCDPEvaluateResponseSchema,
  AddCustomCommandResponse: ModCDPAddCustomCommandResponseSchema,
  AddCustomEventResponse: ModCDPAddCustomEventResponseSchema,
  AddMiddlewareResponse: ModCDPAddMiddlewareResponseSchema,
  ConfigureResponse: ModCDPConfigureResponseSchema,
  PingResponse: ModCDPPingResponseSchema,
  BindingPayload: ModCDPBindingPayloadSchema,
  CustomCommandRegistration: ModCDPCustomCommandRegistrationSchema,
  CustomEventRegistration: ModCDPCustomEventRegistrationSchema,
  MiddlewareRegistration: ModCDPMiddlewareRegistrationSchema,
} as const;
