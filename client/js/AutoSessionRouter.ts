import type { ProtocolParams, ProtocolResult } from "../../types/modcdp.js";

type SendCDP = (method: string, params?: ProtocolParams, session_id?: string | null) => Promise<ProtocolResult>;

export class AutoSessionRouter {
  readonly target_sessions = new Map<string, string>();
  readonly session_targets = new Map<string, Record<string, unknown>>();
  readonly execution_contexts = new Map<string, number>();
  private readonly execution_context_waiters = new Map<string, Set<(context_id: number) => void>>();

  constructor(
    private readonly send: SendCDP,
    private readonly defaultExecutionContextTimeoutMs: () => number,
  ) {}

  sessionIdForTarget(target_id: string) {
    return this.target_sessions.get(target_id) ?? null;
  }

  async attachToTarget(target_id: string) {
    const existing_session_id = this.sessionIdForTarget(target_id);
    if (existing_session_id != null) return existing_session_id;
    const result = await this.send("Target.attachToTarget", { targetId: target_id, flatten: true });
    const session_id = result && typeof result === "object" ? (result as Record<string, unknown>).sessionId : null;
    return typeof session_id === "string" && session_id.length > 0 ? session_id : null;
  }

  recordProtocolEvent(method: string, data: unknown, session_id: string | null) {
    const event_data =
      data && typeof data === "object" && !Array.isArray(data) ? (data as Record<string, unknown>) : {};
    if (method === "Target.attachedToTarget") {
      const attached_session_id = typeof event_data.sessionId === "string" ? event_data.sessionId : session_id;
      const target_info =
        event_data.targetInfo && typeof event_data.targetInfo === "object"
          ? (event_data.targetInfo as Record<string, unknown>)
          : null;
      const target_id = typeof target_info?.targetId === "string" ? target_info.targetId : null;
      if (attached_session_id && target_id && target_info) {
        this.target_sessions.set(target_id, attached_session_id);
        this.session_targets.set(attached_session_id, target_info);
      }
    } else if (method === "Runtime.executionContextCreated") {
      const context = event_data.context && typeof event_data.context === "object" ? event_data.context : null;
      const context_id = context && "id" in context && typeof context.id === "number" ? context.id : null;
      if (session_id && context_id != null) this.recordExecutionContext(session_id, context_id);
    } else if (method === "Target.detachedFromTarget") {
      const detached_session_id = typeof event_data.sessionId === "string" ? event_data.sessionId : session_id;
      if (detached_session_id) this.forgetSession(detached_session_id);
    }
  }

  waitForExecutionContext(session_id: string | null, { timeout_ms }: { timeout_ms?: number } = {}) {
    const effective_timeout_ms = timeout_ms ?? this.defaultExecutionContextTimeoutMs();
    if (!session_id) return Promise.reject(new Error("Cannot wait for a Runtime execution context without a session."));
    const existing = this.execution_contexts.get(session_id);
    if (existing != null) return Promise.resolve(existing);
    return new Promise<number>((resolve, reject) => {
      const timeout = setTimeout(() => {
        const waiters = this.execution_context_waiters.get(session_id);
        waiters?.delete(complete);
        if (waiters?.size === 0) this.execution_context_waiters.delete(session_id);
        reject(new Error(`Timed out waiting for Runtime.executionContextCreated for session ${session_id}.`));
      }, effective_timeout_ms);
      const complete = (context_id: number) => {
        clearTimeout(timeout);
        resolve(context_id);
      };
      const waiters = this.execution_context_waiters.get(session_id);
      if (waiters) waiters.add(complete);
      else this.execution_context_waiters.set(session_id, new Set([complete]));
    });
  }

  private recordExecutionContext(session_id: string, context_id: number) {
    this.execution_contexts.set(session_id, context_id);
    const waiters = this.execution_context_waiters.get(session_id);
    if (!waiters) return;
    this.execution_context_waiters.delete(session_id);
    for (const resolve of waiters) resolve(context_id);
  }

  private forgetSession(session_id: string) {
    const target_info = this.session_targets.get(session_id);
    const target_id = typeof target_info?.targetId === "string" ? target_info.targetId : null;
    if (target_id) this.target_sessions.delete(target_id);
    this.session_targets.delete(session_id);
    this.execution_contexts.delete(session_id);
  }
}
