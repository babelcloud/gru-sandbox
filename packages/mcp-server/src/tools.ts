import { z } from "zod";
import { runPython, runBash, getBoxes } from "./box.js";
import { withLogging } from "./utils.js";

// Common parameters for run tools
export const runToolParams = {
  code: z.string().describe(`The code to run`),
  boxId: z.string().optional()
    .describe(`The ID of an existing box to run the code in.
      If not provided, the system will try to reuse an existing box with matching image.
      The system will first try to use a running box, then a stopped box (which will be started), and finally create a new one if needed.
      Note that without boxId, multiple calls may use different boxes even if they exist.
      If you need to ensure multiple calls use the same box, you must provide a boxId.
      You can get the list of existing boxes by using the list-boxes tool.
      Note: If you run Python code in a box which is created from a non-Python image, you might need to install Python and related tools first using run-bash.
      `),
};

// List boxes handler
export const handleListBoxes = withLogging(
  async (log, {}, { sessionId, signal }) => {
    log({
      level: "info",
      data: `Listing boxes${sessionId ? ` for session: ${sessionId}` : ""}`,
    });
    const boxes = await getBoxes({ signal, sessionId });
    log({ level: "info", data: `Found ${boxes.length} boxes` });
    return {
      content: [
        {
          type: "text" as const,
          text: JSON.stringify(boxes, null, 2),
        },
      ],
    };
  }
);

// Python run tool
export const handleRunPython = withLogging(
  async (log, { boxId, code }, { signal, sessionId }) => {
    log({
      level: "info",
      data: `Executing Python code in box: ${
        boxId || "new box"
      }, sessionId: ${sessionId}`,
    });
    const result = await runPython(code, { signal, sessionId, boxId });
    log({ level: "info", data: "Python code executed successfully" });
    return {
      content: [
        {
          type: "text" as const,
          text: JSON.stringify(result, null, 2),
        },
      ],
    };
  }
);

// Bash run tool
export const handleRunBash = withLogging(
  async (log, { boxId, code }, { signal, sessionId }) => {
    log({
      level: "info",
      data: `Executing Bash command in box ${boxId || "new box"}`,
    });
    const result = await runBash(code, { signal, sessionId, boxId });
    log({ level: "info", data: "Bash command executed successfully" });
    return {
      content: [
        {
          type: "text" as const,
          text: JSON.stringify(result, null, 2),
        },
      ],
    };
  }
);
