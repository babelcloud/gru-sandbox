import { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";
import { StdioServerTransport } from "@modelcontextprotocol/sdk/server/stdio.js";
import { boxTemplate, handleBoxResource } from "./resources.js";
import {
  runToolParams,
  handleRunPython,
  handleRunBash,
  handleListBoxes,
} from "./tools.js";
import type { LoggingMessageNotification } from "@modelcontextprotocol/sdk/types.js";
import type { RequestHandlerExtra } from "@modelcontextprotocol/sdk/shared/protocol.js";

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
const log = async (
  params: LoggingMessageNotification["params"]
): Promise<void> => {
  if (enableLogging) {
    await mcpServer.server.sendLoggingMessage(params);
  }
};

// Register box resource
mcpServer.resource("box", boxTemplate(log), handleBoxResource(log));

mcpServer.prompt(
  "run-python",
  "Run Python code in a sandbox.",
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

mcpServer.tool("list-boxes", "List all boxes.", {}, handleListBoxes(log));

// Register run tools with descriptions
mcpServer.tool(
  "run-python",
  `Run Python code in a sandbox. 
If no boxId is provided, the system will try to reuse an existing box with matching image. 
The system will first try to use a running box, then a stopped box (which will be started), and finally create a new one if needed. 
Note that without boxId, multiple calls may use different boxes even if they exist. 
If you need to ensure multiple calls use the same box, you must provide a boxId. 
The Python image comes with uv package manager pre-installed and pip is not available. 
Please use uv for installing Python packages.`,
  runToolParams,
  handleRunPython(log)
);

mcpServer.tool(
  "run-bash",
  `Run Bash commands in a sandbox. 
If no boxId is provided, the system will try to reuse an existing box with matching image. 
The system will first try to use a running box, then a stopped box (which will be started), and finally create a new one if needed. 
Note that without boxId, multiple calls may use different boxes even if they exist. 
If you need to ensure multiple calls use the same box, you must provide a boxId.`,
  runToolParams,
  handleRunBash(log)
);

// Start server
const transport = new StdioServerTransport();
await mcpServer.connect(transport);
// Log successful startup
log({
  level: "info",
  data: "Server started successfully",
});
