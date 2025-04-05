import { z } from "zod";
import { withLogging } from "../utils.js";
import TurndownService from "turndown";
import type { LogFunction } from "../utils.js";
import { executeBashCommand } from "./run-bash.js";
import { Box } from "../sdk/types.js";
import { readFileHandler } from "./read-file.js";

export const BROWSER_OPEN_URL_TOOL = "browser-open-url";
export const BROWSER_OPEN_URL_DESCRIPTION = "Fetch and view content from a URL, optionally converting HTML to markdown. Usually helpful for reading web pages.";

export const browserOpenUrlParams = {
  url: z.string().describe("The URL to fetch content from (must start with http:// or https://)"),
  contentType: z.enum(["markdown", "html"]).default("markdown")
    .describe("The desired output format: 'markdown' or 'html' (defaults to 'markdown')"),
};

const browserOpenUrlParamsShape = z.object(browserOpenUrlParams);

export const handleBrowserOpenUrl = withLogging(async (log: LogFunction, params: z.infer<typeof browserOpenUrlParamsShape>, { signal }: { signal?: AbortSignal }) => {
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
    // Execute the browser_open_url command
    log({ level: "info", data: `Fetching content from URL: ${url}` });
    const bashResult = await executeBashCommand(log, {
      code: `browser_open_url -screenshot ${url}`,
    }, { signal: undefined, sessionId: undefined, stdoutLineLimit: 100, stderrLineLimit: 100 });

    const commandResult = JSON.parse(bashResult.content[0].text);
    
    // Get text content (HTML or markdown)
    const textContent = await getTextReturnContent(commandResult, contentType, log, signal);
    
    // Get image content (screenshot)
    const imageContent = await getImageReturnContent(commandResult, log, signal);
    
    // Prepare content array
    const content = [];
    
    // Add image content if available
    if (imageContent) {
      content.push(...imageContent);
    }
    
    // Add text content
    content.push({
      type: "text" as const,
      text: textContent,
    });
    
    return { content };

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

async function getTextReturnContent(commandResult: { stdout: string; stderr: string; box: Box }, contentType: "markdown" | "html", log: LogFunction, signal?: AbortSignal): Promise<string> {
  // try to parse the stdout
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
  
  // Convert HTML to markdown
  const turndownService = new TurndownService();
  return turndownService.turndown(htmlContent);
}

async function getImageReturnContent(commandResult: { stdout: string; stderr: string; box: Box }, log: LogFunction, signal?: AbortSignal): Promise<{ type: "image"; data: string; mimeType: string }[] | null> {
  // try to parse the stdout to get screenshot path
  var screenshotPaths = [];
  try {
    const result = JSON.parse(commandResult.stdout);
    screenshotPaths = result.screenshot_files;
  } catch (error) {
    log({ level: "error", data: `Error parsing stdout for screenshot: ${error}` });
    return null;
  }

  if (!screenshotPaths.length) {
    return null;
  }

  const images = [];
  try {
    for (const screenshotPath of screenshotPaths) {
      const screenshotResult = await readFileHandler(log, { path: screenshotPath, boxId: commandResult.box.id }, { signal });
      
      // Check if we got an image back
      if (screenshotResult.content[0].type === "image") {
        images.push({
          type: "image" as const,
          data: screenshotResult.content[0].data,
          mimeType: screenshotResult.content[0].mimeType,
        });
      } else if (screenshotResult.content[0].type === "text" && screenshotResult.content[0].text.startsWith("http")) {
          // If the screenshot is too large, we get a URL instead
          // In this case, we can't return it as an image type, so return null
          log({ level: "info", data: `Screenshot is too large, URL: ${screenshotResult.content[0].text}` });
          return null;
      }
    }
  } catch (error) {
    log({ level: "error", data: `Error reading screenshot: ${error}` });
  }

  if (!images.length) {
    return null;
  }

  return images;
}
