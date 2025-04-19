import { withLogging } from "../utils.js";
import { config } from "../config.js";
import { GBox } from "../sdk/index.js";
import type { Logger } from "../sdk/types";

export const LIST_BOXES_TOOL = "list-boxes";
export const LIST_BOXES_DESCRIPTION = "List all boxes.";

export const handleListBoxes = withLogging(
  async (logger: Logger, {}, { sessionId, signal }) => {
    const gbox = new GBox({
      apiUrl: config.apiServer.url,
      logger,
    });

    logger.info(`Listing boxes${sessionId ? ` for session: ${sessionId}` : ""}`);

    const response = await gbox.box.getBoxes({ signal, sessionId });

    logger.info(`Found ${response.count} boxes`);

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
