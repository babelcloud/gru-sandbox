import { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";
import { StdioServerTransport } from "@modelcontextprotocol/sdk/server/stdio.js";
import { boxTemplate, handleBoxResource } from "./resources.js";
import {
  LIST_BOXES_TOOL,
  LIST_BOXES_DESCRIPTION,
  RUN_PYTHON_TOOL,
  RUN_PYTHON_DESCRIPTION,
  RUN_BASH_TOOL,
  RUN_BASH_DESCRIPTION,
  READ_FILE_TOOL,
  READ_FILE_DESCRIPTION,
  handleListBoxes,
  handleRunPython,
  handleRunBash,
  handleReadFile,
  runPythonParams,
  readFileParams,
  handleViewByUrl,
  viewByUrlParams,
  VIEW_BY_URL_TOOL,
  VIEW_BY_URL_DESCRIPTION,
  runBashParams,
} from "./tools/index.js";
import type { LoggingMessageNotification } from "@modelcontextprotocol/sdk/types.js";
import type { RequestHandlerExtra } from "@modelcontextprotocol/sdk/shared/protocol.js";
import type { LogFn } from "./types.js";
import {
  handleRunTypescript,
  RUN_TYPESCRIPT_DESCRIPTION,
  RUN_TYPESCRIPT_TOOL,
  runTypescriptParams,
} from "./tools/run-typescript.js";

const enableLogging = true;

// Create MCP server instance
const mcpServer = new McpServer(
  {
    name: "gbox-mcp-server",
    version: "1.0.0",
  },
  {
    capabilities: {
      prompts: {},
      resources: {},
      tools: {},
      ...(enableLogging ? { logging: {} } : {}),
    },
  }
);
const log: LogFn = async (
  params: LoggingMessageNotification["params"]
): Promise<void> => {
  if (enableLogging) {
    await mcpServer.server.sendLoggingMessage(params);
  }
};

// Register box resource
mcpServer.resource("box", boxTemplate(log), handleBoxResource(log));

// Register run-python prompt
mcpServer.prompt(
  RUN_PYTHON_TOOL,
  RUN_PYTHON_DESCRIPTION,
  (_: RequestHandlerExtra) => {
    return {
      messages: [
        {
          role: "user",
          content: {
            type: "text",
            text: `When running Python code in a sandbox, if you don't provide a boxId, the system will try to reuse an existing box with matching image. 
              The system will first try to use a running box, then a stopped box (which will be started), and finally create a new one if needed. 
              Note that without boxId, multiple calls may use different boxes even if they exist. 
              If you need to ensure multiple calls use the same box, you must provide a boxId. 
              You can use the 'list-boxes' tool to see available boxes.`,
          },
        },
      ],
    };
  }
);

// Register tools
mcpServer.tool(
  LIST_BOXES_TOOL,
  LIST_BOXES_DESCRIPTION,
  {},
  handleListBoxes(log)
);

mcpServer.tool(
  READ_FILE_TOOL,
  READ_FILE_DESCRIPTION,
  readFileParams,
  handleReadFile(log)
);

mcpServer.tool(
  RUN_PYTHON_TOOL,
  RUN_PYTHON_DESCRIPTION,
  runPythonParams,
  handleRunPython(log)
);

mcpServer.tool(
  RUN_TYPESCRIPT_TOOL,
  RUN_TYPESCRIPT_DESCRIPTION,
  runTypescriptParams,
  handleRunTypescript(log)
);

mcpServer.tool(
  RUN_BASH_TOOL,
  RUN_BASH_DESCRIPTION,
  runBashParams,
  handleRunBash(log)
);

mcpServer.tool(
  VIEW_BY_URL_TOOL,
  VIEW_BY_URL_DESCRIPTION,
  viewByUrlParams,
  handleViewByUrl(log)
);

// Start server
const transport = new StdioServerTransport();
await mcpServer.connect(transport);
// Log successful startup
log({
  level: "info",
  data: "Server started successfully",
});
