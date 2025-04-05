import { z } from "zod";
import { withLogging } from "../utils.js";
import TurndownService from "turndown";
import type { LogFunction } from "../utils.js";
import { executeBashCommand } from "./run-bash.js";
import { Box } from "../sdk/types.js";
import { readFileHandler } from "./read-file.js";

export const VIEW_BY_URL_TOOL = "view-by-url";
export const VIEW_BY_URL_DESCRIPTION = "Fetch and view content from a URL, optionally converting HTML to markdown. Usually helpful for reading web pages.";

export const viewByUrlParams = {
  url: z.string().describe("The URL to fetch content from (must start with http:// or https://)"),
  contentType: z.enum(["markdown", "html"]).default("markdown")
    .describe("The desired output format: 'markdown' or 'html' (defaults to 'markdown')"),
};

const viewByUrlParamsShape = z.object(viewByUrlParams);

export const handleViewByUrl = withLogging(async (log: LogFunction, params: z.infer<typeof viewByUrlParamsShape>, { signal }: { signal?: AbortSignal }) => {
  const { url, contentType = "markdown" } = params;

  // Validate URL format
  if (!url.startsWith("http://") && !url.startsWith("https://")) {
    return {
      content: [
        {
          type: "text" as const,
          text: "Error: URL must start with http:// or https://",
        },
      ],
    };
  }

  try {
    // Execute the view_by_url command
    log({ level: "info", data: `Fetching content from URL: ${url}` });
    const bashResult = await executeBashCommand(log, {
      code: `view_by_url ${url}`,
    }, { signal: undefined, sessionId: undefined, stdoutLineLimit: 100, stderrLineLimit: 100 });

    const commandResult = JSON.parse(bashResult.content[0].text);
    const content = await getReturnContent(commandResult, contentType, log, signal);
    return {
      content: [
        {
          type: "text" as const,
          text: content,
        },
      ],
    };

  } catch (error) {
    const errorMessage = error instanceof Error ? error.message : "An unknown error occurred";
    return {
      content: [
        {
          type: "text" as const,
          text: `Error: ${errorMessage}`,
        },
      ],
    };
  }
});

async function getReturnContent(commandResult: { stdout: string; stderr: string; box: Box }, contentType: "markdown" | "html", log: LogFunction, signal?: AbortSignal): Promise<string> {
  // try to parse the sdtout
  var path = "";
  try {
    const result = JSON.parse(commandResult.stdout);
    path = result.output_file;
    log({ level: "debug", data: result });
  } catch (error) {
    log({ level: "error", data: `Error parsing stdout: ${error}` });
  }

  if (!path){
    return commandResult.stdout;
  }

  
  const readFileResult = await readFileHandler(log, { path, boxId: commandResult.box.id }, { signal });

  const htmlContent = readFileResult.content[0].type === "text" ? readFileResult.content[0].text : null;

  if (!htmlContent){
    return commandResult.stdout;
  }

  if (htmlContent.startsWith("http")){
    return "[File is too large to display]";
  }

  if (contentType === "html") {
    return htmlContent;
  }
  const turndownService = new TurndownService();
  return turndownService.turndown(htmlContent);
}
