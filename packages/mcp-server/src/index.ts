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
  handleBrowserOpenUrl,
  browserOpenUrlParams,
  BROWSER_OPEN_URL_TOOL,
  BROWSER_OPEN_URL_DESCRIPTION,
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
//mcpServer.resource("box", boxTemplate(log), handleBoxResource(log));


const GBOX_MANUAL = "gbox-manual";
const GBOX_MANUAL_DESCRIPTION = "A manual for the gbox command line tool.";
const GBOX_MANUAL_CONTENT = `
# GBox Manual

## Overview
Gbox is a set of tools that allows you to complete various tasks. All the tools are executed in a sandboxed environment. Gbox is developed by Gru AI.

## Usage
### run-python
If you need to execute a standalone python script, you can use the run-python tool. 

### run-bash
If you need to execute a standalone bash script, you can use the run-bash tool. 

### read-file
If you need to read a file from the sandbox, you can use the read-file tool. 

### browser-open-url
If you need to view a web page/PDF/etc, you can use the browser-open-url tool. 

### list-boxes
If you need to execute tools in different boxes, you can use the list-boxes tool to list all the boxes and pick the one you want.
`
// Register run-python prompt
mcpServer.prompt(
  GBOX_MANUAL,
  GBOX_MANUAL_DESCRIPTION,
  (_: RequestHandlerExtra) => {
    return {
      messages: [
        {
          role: "user",
          content: {
            type: "text",
            text: GBOX_MANUAL_CONTENT
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
  BROWSER_OPEN_URL_TOOL,
  BROWSER_OPEN_URL_DESCRIPTION,
  browserOpenUrlParams,
  handleBrowserOpenUrl(log)
);

// Start server
const transport = new StdioServerTransport();
await mcpServer.connect(transport);
// Log successful startup
log({
  level: "info",
  data: "Server started successfully",
});
