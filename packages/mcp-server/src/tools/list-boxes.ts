import { withLogging } from "../utils.js";
import { config } from "../config.js";
import { GBox } from "../sdk/index.js";
import { MCPLogger } from "../mcp-logger.js";

export const LIST_BOXES_TOOL = "list-boxes";
export const LIST_BOXES_DESCRIPTION = "List all boxes.";

export const handleListBoxes = withLogging(
  async (log, {}, { sessionId, signal }) => {
    const logger = new MCPLogger(log);
    const gbox = new GBox({
      apiUrl: config.apiServer.url,
      logger,
    });

    log({
      level: "info",
      data: `Listing boxes${sessionId ? ` for session: ${sessionId}` : ""}`,
    });

    const response = await gbox.box.getBoxes({ signal, sessionId });

    log({ level: "info", data: `Found ${response.boxes.length} boxes` });

    return {
      content: [
        {
          type: "text" as const,
          text: JSON.stringify(response.boxes, null, 2),
        },
      ],
    };
  }
);
