import { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";
import { boxTemplate, handleBoxResource } from "./resources.js";
import {
  RUN_PYTHON_TOOL,
  RUN_PYTHON_DESCRIPTION,
  RUN_BASH_TOOL,
  RUN_BASH_DESCRIPTION,
  READ_FILE_TOOL,
  READ_FILE_DESCRIPTION,
  WRITE_FILE_TOOL,
  WRITE_FILE_DESCRIPTION,
  LIST_BOXES_TOOL,
  LIST_BOXES_DESCRIPTION,
  handleRunPython,
  handleRunBash,
  handleReadFile,
  runPythonParams,
  runBashParams,
  readFileParams,
  handleWriteFile,
  writeFileParams,
  handleListBoxes,
  VIEW_URL_AS_TOOL,
  VIEW_URL_AS_DESCRIPTION,
  viewUrlAsParams,
  handleViewUrlAs,
} from "./tools/index.js";
import type { RequestHandlerExtra } from "@modelcontextprotocol/sdk/shared/protocol.js";
import express from "express";
import { SSEServerTransport } from "@modelcontextprotocol/sdk/server/sse.js";
import { MCPLogger } from "./mcp-logger.js";

let transport: SSEServerTransport | null = null;

const app = express();

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
    },
  }
);

// Create an instance of MCPLogger
const logger = new MCPLogger();

// Register box resource, passing the logger instance
mcpServer.resource("box", boxTemplate(logger), handleBoxResource(logger));

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

### write-file
If you need to write a file to the sandbox, you can use the write-file tool. 
`

// Register gbox manual prompt
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
  handleListBoxes(logger)
);

mcpServer.tool(
  READ_FILE_TOOL,
  READ_FILE_DESCRIPTION,
  readFileParams,
  handleReadFile(logger)
);

mcpServer.tool(
  WRITE_FILE_TOOL,
  WRITE_FILE_DESCRIPTION,
  writeFileParams,
  handleWriteFile(logger)
);

mcpServer.tool(
  RUN_PYTHON_TOOL,
  RUN_PYTHON_DESCRIPTION,
  runPythonParams,
  handleRunPython(logger)
);

mcpServer.tool(
  RUN_BASH_TOOL,
  RUN_BASH_DESCRIPTION,
  runBashParams,
  handleRunBash(logger)
);

mcpServer.tool(
  VIEW_URL_AS_TOOL,
  VIEW_URL_AS_DESCRIPTION,
  viewUrlAsParams,
  handleViewUrlAs(logger)
);
app.get("/sse", (req, res) => {
  transport = new SSEServerTransport("/messages", res);
  mcpServer.connect(transport);
});

app.post("/messages", (req, res) => {
  if (transport) {
    transport.handlePostMessage(req, res);
  }
});

const port = process.env.PORT || 28090;

app.listen(port, () => {
  // console.log(`Server started successfully on port ${port}`);
});
