import { withLogging } from "../utils.js";
import { config } from "../config.js";
import { GBox } from "../sdk/index.js";
import type { Logger } from "../sdk/types";
import { z } from "zod";

export const WRITE_FILE_TOOL = "write-file";
export const WRITE_FILE_DESCRIPTION = `Write content to a file in the sandbox.
If the file doesn't exist, it will be created. If it exists, it will be overwritten.
The path must start with / and requires a boxId to specify which box to write to.`;

export const writeFileParams = {
  path: z
    .string()
    .describe(`The path to the file in the sandbox. Must start with /`),
  content: z.string().describe(`The content to write to the file.`),
  boxId: z.string().describe(`The ID of the box to write the file to.`),
};

export const handleWriteFile = withLogging(
  async (logger: Logger, { path, content, boxId }, { signal }) => {
    const gbox = new GBox({
      apiUrl: config.apiServer.url,
      logger,
    });

    logger.info(`Writing to file: ${path} from box: ${boxId}`);

    const writeResponse = await gbox.file.writeFile(
      boxId,
      path,
      content,
      signal
    );
    if (!writeResponse || !writeResponse.success) {
      return {
        content: [
          {
            type: "text" as const,
            text: writeResponse?.message || "Failed to write file",
          },
        ],
      };
    }

    return {
      content: [
        {
          type: "text" as const,
          text: "File written successfully",
        },
      ],
    };
  }
);
