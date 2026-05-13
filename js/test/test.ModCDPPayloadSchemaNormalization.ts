import assert from "node:assert/strict";
import { test } from "vitest";
import { z } from "zod";

import { normalizeModCDPPayloadSchema } from "../src/types/modcdp.js";

test("payload schema normalization accepts empty zod shapes", () => {
  const schema = normalizeModCDPPayloadSchema({});
  assert.deepEqual(schema?.parse({ value: 1 }), { value: 1 });
});

test("payload schema normalization rejects unsupported schema specs", () => {
  assert.throws(() => normalizeModCDPPayloadSchema("not-a-schema" as never), /Unsupported payload schema/);
});

test("payload schema normalization accepts non-empty zod shapes", () => {
  const schema = normalizeModCDPPayloadSchema({ value: z.string() });
  assert.deepEqual(schema?.parse({ value: "ok", extra: true }), { value: "ok", extra: true });
});
