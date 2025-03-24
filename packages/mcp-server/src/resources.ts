import {
  ResourceTemplate,
  ReadResourceTemplateCallback,
} from "@modelcontextprotocol/sdk/server/mcp.js";
import { RequestHandlerExtra } from "@modelcontextprotocol/sdk/shared/protocol.js";
import { Variables } from "@modelcontextprotocol/sdk/shared/uriTemplate.js";
import { getBoxes, getBox } from "./box.js";
import {
  withLogging,
  withLoggingResourceTemplate,
  type LogFunction,
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

    const boxes = await getBoxes({ signal, sessionId });
    log({ level: "info", data: `Found ${boxes?.length || 0} boxes` });

    if (!boxes || boxes.length === 0) {
      log({ level: "info", data: "No boxes found, returning empty list" });
      return { resources: [] };
    }

    log({ level: "info", data: "Mapping boxes to resource format" });
    const resources = boxes.map((box: Box) => {
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

    const box = await getBox(boxId, { signal, sessionId });
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
