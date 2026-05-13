import { defineConfig } from "vitest/config";

export default defineConfig({
  test: {
    include: ["js/test/**/*.test.ts", "js/test/test.*.ts"],
    exclude: ["js/test/helpers.*.ts"],
    environment: "node",
  },
});
