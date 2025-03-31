import { LogFunction, withLogging } from "../utils.js";
import { config } from "../config.js";
import { GBox, FILE_SIZE_LIMITS } from "../sdk/index.js";
import { MCPLogger } from "../mcp-logger.js";
import { z } from "zod";
import type { FileStat } from "../sdk/types.js";

export const READ_FILE_TOOL = "read-file";
export const READ_FILE_DESCRIPTION = `Read a file from the API server.
If the file is a small text file (less than 1MB), the content will be returned directly.
If the file is a small image or audio file (less than 5MB), the content will be returned as base64 encoded data.
For large files or unsupported types, a resource URL will be returned with file metadata.
The path must start with / and requires a boxId to specify which box to read from.`;

export const readFileParams = {
  path: z
    .string()
    .describe(`The path to the file in the share directory. Must start with /`),
  boxId: z.string().describe(`The ID of the box to read the file from.`),
};

// Helper function to handle file content based on type and size
async function handleFileContent(
  sdk: GBox,
  fileStat: FileStat,
  mimeType: string,
  contentLength: number,
  path: string,
  boxId: string,
  logger: MCPLogger,
  signal?: AbortSignal,
) {
  // For directories, return error
  if (fileStat.type === "directory") {
    return {
      content: [
        {
          type: "text" as const,
          text: "Cannot read directory content",
        },
      ],
      isError: true,
    };
  }

  // For small text files (less than 1MB), return content directly
  if (mimeType.startsWith("text/") && contentLength < FILE_SIZE_LIMITS.TEXT) {
    const text = await sdk.file.readFileAsText(`${boxId}${path}`, signal);
    return {
      content: [
        {
          type: "text" as const,
          text: text || "",
        },
      ],
    };
  }

  // For small images (less than 5MB), return base64 encoded content
  if (
    mimeType.startsWith("image/") &&
    contentLength < FILE_SIZE_LIMITS.BINARY
  ) {
    logger.info(`Reading image file: ${path}, from box: ${boxId}`);
    const buffer = await sdk.file.readFileAsBuffer(`${boxId}${path}`, signal);
    if (!buffer) {
      return {
        content: [
          {
            type: "text" as const,
            text: "Failed to read image file",
          },
        ],
        isError: true,
      };
    }
    const base64 = sdk.file.bufferToBase64(buffer);
    logger.info("base64: ", base64);
    return {
      content: [
        {
          type: "image" as const,
          data: base64,
          mimeType,
        },
      ],
    };
  }

  // For small audio files (less than 5MB), return base64 encoded content
  if (
    mimeType.startsWith("audio/") &&
    contentLength < FILE_SIZE_LIMITS.BINARY
  ) {
    const buffer = await sdk.file.readFileAsBuffer(`${boxId}${path}`, signal);
    if (!buffer) {
      return {
        content: [
          {
            type: "text" as const,
            text: "Failed to read audio file",
          },
        ],
        isError: true,
      };
    }
    const base64 = sdk.file.bufferToBase64(buffer);
    return {
      content: [
        {
          type: "audio" as const,
          data: base64,
          mimeType,
        },
      ],
    };
  }

  // For large files or unsupported types, return resource URI
  return {
    content: [
      {
        type: "text" as const,
        text: `${config.apiServer.url}/files/${boxId}${path}`,
      },
    ],
  };
}

const readFileParamsShape = z.object(readFileParams);
// Read file handler implementation
export async function readFileHandler(
  log: LogFunction,
  { path, boxId }: z.infer<typeof readFileParamsShape>,
  { signal }: { signal?: AbortSignal }
) {
  const logger = new MCPLogger(log);
  const gbox = new GBox({
    apiUrl: config.apiServer.url,
    logger,
  });

  logger.info(`Reading file: ${path}${boxId ? ` from box: ${boxId}` : ""}`);

  // First check if file exists and get metadata
  let metadata = await gbox.file.getFileMetadata(`${boxId}${path}`, signal);

  // If file doesn't exist and we have a boxId, try to share it
  if (!metadata && boxId) {
    logger.info(`File not found, attempting to share from box: ${boxId}`);

    // Share file from box
    const shareResponse = await gbox.file.shareFile(path, boxId, signal);
    if (!shareResponse || !shareResponse.success) {
      return {
        content: [
          {
            type: "text" as const,
            text: shareResponse?.message || "Failed to share file",
          },
        ],
        isError: true,
      };
    }

    logger.info(
      `File shared successfully: ${
        shareResponse.message
      }\nShared files: ${shareResponse.fileList
        .map((f) => f.path)
        .join(", ")}`
    );

    // Retry getting file metadata after sharing
    metadata = await gbox.file.getFileMetadata(`${boxId}${path}`, signal);
  }

  if (!metadata) {
    return {
      content: [
        {
          type: "text" as const,
          text: "File not found",
        },
      ],
      isError: true,
    };
  }

  const { fileStat, mimeType, contentLength } = metadata;

  logger.info(
    `File metadata: ${JSON.stringify(
      fileStat
    )}, mimeType: ${mimeType}, contentLength: ${contentLength}`
  );

  return handleFileContent(
    gbox,
    fileStat,
    mimeType,
    contentLength,
    path,
    boxId,
    logger,
    signal
  );
}

export const handleReadFile = withLogging(readFileHandler);
