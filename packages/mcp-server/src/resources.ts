import { config } from "./config.js";
import { MCPLogger } from "./mcp-logger.js";
import { GBox } from "./sdk/index.js";
import {
  type LogFunction,
  withLogging,
  withLoggingResourceTemplate,
} from "./utils.js";

// Box interface
interface Box {
  id: string;
  image: string;
  status: string;
}

// Define box resource template
const boxTemplate = withLoggingResourceTemplate("gbox:///boxes/{boxId}", {
  list: async (log: LogFunction, { signal, sessionId }) => {
    log({ level: "info", data: "Starting to fetch boxes" });

    const logger = new MCPLogger(log);
    const sdk = new GBox({
      apiUrl: config.apiServer.url,
      logger,
    });

    const response = await sdk.box.getBoxes({ signal, sessionId });
    log({ level: "info", data: `Found ${response.boxes.length} boxes` });

    if (!response.boxes || response.boxes.length === 0) {
      log({ level: "info", data: "No boxes found, returning empty list" });
      return { resources: [] };
    }

    log({ level: "info", data: "Mapping boxes to resource format" });
    const resources = response.boxes.map((box: Box) => {
      const resource = {
        uri: `gbox:///boxes/${box.id}`,
        name: `Box ${box.id}`,
        description: `Status: ${box.status}, Image: ${box.image}. Note: When executing code, if the box is stopped, it will be automatically started first. This is suitable when previous processes have stopped but disk contents are preserved.`,
        mimeType: "application/json",
      };
      log({ level: "debug", data: `Mapped box ${box.id} to resource` });
      return resource;
    });

    log({ level: "info", data: "Successfully completed box list operation" });
    return { resources };
  },
});

// Box resource handler
export const handleBoxResource = withLogging(
  async (log, uri, { boxId }, { signal, sessionId }) => {
    boxId = Array.isArray(boxId) ? boxId[0] : boxId;
    if (!boxId) {
      log({ level: "error", data: "Box ID is missing" });
      return {
        contents: [
          {
            uri: uri.href,
            mimeType: "application/json",
            text: JSON.stringify({ error: "Box ID is required" }, null, 2),
          },
        ],
      };
    }

    const logger = new MCPLogger(log);
    const sdk = new GBox({
      apiUrl: config.apiServer.url,
      logger,
    });

    const box = await sdk.box.getBox(boxId, { signal, sessionId });
    log({ level: "info", data: `Successfully fetched box ${boxId}` });
    return {
      contents: [
        {
          uri: uri.href,
          mimeType: "application/json",
          text: JSON.stringify(
            {
              ...box,
              description:
                "When executing code, if the box is stopped, it will be automatically started first. This is suitable when previous processes have stopped but disk contents are preserved.",
            },
            null,
            2
          ),
        },
      ],
    };
  }
);

export { boxTemplate };
