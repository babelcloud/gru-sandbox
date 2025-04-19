export * from "./list-boxes.js";
export * from "./run-python.js";
export * from "./run-bash.js";
export * from "./read-file.js";
export * from "./write-file.js";
export * from "./view-url-as.js";

// Re-export tool names and descriptions for convenience
export const TOOL_NAMES = {
  LIST_BOXES: "list-boxes",
  RUN_PYTHON: "run-python",
  RUN_BASH: "run-bash",
  READ_FILE: "read-file",
  WRITE_FILE: "write-file",
  VIEW_URL_AS: "view-url-as",
} as const;

export type ToolName = (typeof TOOL_NAMES)[keyof typeof TOOL_NAMES];
