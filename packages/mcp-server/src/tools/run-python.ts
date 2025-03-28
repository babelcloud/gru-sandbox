import { withLogging } from "../utils.js";
import { config } from "../config.js";
import { GBox } from "../sdk/index.js";
import { MCPLogger } from "../mcp-logger.js";
import { z } from "zod";

export const RUN_PYTHON_TOOL = "run-python";
export const RUN_PYTHON_DESCRIPTION = `Run Python code in a sandbox. 
If no boxId is provided, the system will try to reuse an existing box with matching image. 
The system will first try to use a running box, then a stopped box (which will be started), and finally create a new one if needed. 
Note that without boxId, multiple calls may use different boxes even if they exist. 
If you need to ensure multiple calls use the same box, you must provide a boxId. 
The Python image comes with uv package manager pre-installed and pip is not available. 
The following Python packages are pre-installed: numpy, scipy, pandas, scikit-learn, requests, beautifulsoup4, pillow.
To install additional Python packages, use run-bash tool to execute 'uv pip install --system' as virtual environments are not yet supported.

To persist files after sandbox reclamation, save them to /var/gbox/share directory. 
Files in this directory will be retained for a period of time after the sandbox is reclaimed.

To read files generated by your program, use the read-file tool with the boxId returned from this tool.`;

export const runPythonParams = {
  code: z.string().describe(`The Python code to run`),
  boxId: z.string().optional()
    .describe(`The ID of an existing box to run the code in.
      If not provided, the system will try to reuse an existing box with matching image.
      The system will first try to use a running box, then a stopped box (which will be started), and finally create a new one if needed.
      Note that without boxId, multiple calls may use different boxes even if they exist.
      If you need to ensure multiple calls use the same box, you must provide a boxId.
      You can get the list of existing boxes by using the list-boxes tool.
      Note: If you run Python code in a box which is created from a non-Python image, you might need to install Python and related tools first using run-bash.
      `),
};

export const handleRunPython = withLogging(
  async (log, { boxId, code }, { signal, sessionId }) => {
    const logger = new MCPLogger(log);
    const gbox = new GBox({
      apiUrl: config.apiServer.url,
      logger,
    });

    logger.info(
      `Executing Python code in box: ${boxId || "new box"} ${
        sessionId ? `for session: ${sessionId}` : ""
      }`
    );

    // Get or create box
    const id = await gbox.box.getOrCreateBox({
      boxId,
      image: config.images.python,
      sessionId,
      signal,
    });

    // Run command
    const result = await gbox.box.runInBox(
      id,
      ["python3"],
      code,
      100, // stdoutLineLimit
      100, // stderrLineLimit
      { signal, sessionId }
    );

    log({ level: "info", data: "Python code executed successfully" });
    return {
      content: [
        {
          type: "text" as const,
          text: JSON.stringify(result, null, 2),
        },
      ],
    };
  }
);
