import { z } from "zod";

import type { ProtocolResult } from "../types/modcdp.js";

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
      .custom<Awaited<ReturnType<import("../launcher/BrowserLauncher.js").BrowserLauncher["launch"]>>>()
      .nullable(),
    launch_promise: z
      .promise(z.custom<Awaited<ReturnType<import("../launcher/BrowserLauncher.js").BrowserLauncher["launch"]>>>())
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
