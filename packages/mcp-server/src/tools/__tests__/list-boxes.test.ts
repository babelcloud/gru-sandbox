import { describe, it, expect, vi } from "vitest";
import { handleListBoxes } from "../index";
import type { Logger } from "../../sdk/types";

describe("handleListBoxes", () => {
  it("should return a list of boxes", async () => {
    const mockLogger: Logger = {
      debug: vi.fn(),
      info: vi.fn(),
      warn: vi.fn(),
      error: vi.fn(),
    };

    const handler = handleListBoxes(mockLogger);

    const result = await handler(
      {},
      {},
      {
        sessionId: "test-session",
        signal: new AbortController().signal,
      }
    );

    expect(result).toBeDefined();
    expect(result.content).toBeDefined();
    expect(Array.isArray(result.content)).toBe(true);
    expect(result.content[0].type).toBe("text");
  });
});
