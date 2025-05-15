import { z } from "zod";// Import type guard if needed
import { withLogging } from "../utils.js"; // Added import
import { config } from "../config.js"; // Added import
import { Gbox } from "../service/index.js";
import { BrowserPage, BrowserContext, BoxBrowserManager, type VisionScreenshotResult, type VisionScreenshotParams } from "../service/gbox.instance.js";
import type { Logger } from '../mcp-logger.js';

// Define the Zod schema for the tool's parameters, including name and description
export const ViewUrlAsSchema = z.object({
  name: z.literal("view-url-as"),
  description: z
    .string()
    .default(
      "Fetch and view content from a URL as HTML, Markdown, or screenshot. Optionally specify a boxId to reuse a specific browser session."
    )
    .describe(
      // Using .describe() for a potentially longer description if needed elsewhere
      "Navigates to a URL in a browser context and returns its content as HTML, Markdown, or takes a screenshot. Use the boxId parameter to maintain session state across calls."
    ),
  parameters: z.object({
    url: z.string().url({
      message:
        "The URL to fetch content from (must start with http:// or https://).",
    }),
    as: z.enum(["html", "markdown", "screenshot"], {
      errorMap: () => ({
        message:
          "The desired output format: 'html', 'markdown', or 'screenshot'.",
      }),
    }),
    boxId: z.preprocess(
      // Preprocess to handle null input
      (val) => (val === null ? undefined : val),
      z.string().optional() // Keep the original validation
    ) // Added boxId
      .describe(`The ID of an existing browser box to use.
        If not provided, the system will try to reuse an existing box with a browser image.
        The system will first try to use a running box, then a stopped box (which will be started), and finally create a new one if needed.
        Note that without boxId, multiple calls may use different boxes even if they exist.
        If you need to ensure multiple calls use the same box, you must provide a boxId.
        You can get the list of existing boxes by using the list-boxes tool.
        `),
  }),
});

// Derive constants from the schema
export const VIEW_URL_AS_TOOL = ViewUrlAsSchema.shape.name.value;
export const VIEW_URL_AS_DESCRIPTION =
  ViewUrlAsSchema.shape.description._def.defaultValue(); // Get the default value
export const viewUrlAsParams = ViewUrlAsSchema.shape.parameters.shape; // Use .shape to get the raw shape for McpServer.tool

// Define the expected content structure(s)
type TextContent = { type: "text"; text: string };
type ImageContent = { type: "image"; data: string; mimeType: string };
type ErrorContent = { type: "error"; text: string };
type ResultContent = TextContent | ImageContent | ErrorContent;

// Infer the type for the parameters object from the schema
type ViewUrlAsParams = z.infer<typeof ViewUrlAsSchema.shape.parameters>;

// Define the internal action function for the tool
async function _handleViewUrlAs(
  logger: Logger, // Logger comes first from withLogging
  { url, as, boxId }: ViewUrlAsParams, // Destructure params
  { signal, sessionId }: { signal: AbortSignal; sessionId?: string }
): Promise<{ content: Array<ResultContent> }> {
  // UPDATED return type
  const gbox = new Gbox();

  let contextId: string | null = null;
  let pageId: string | null = null;
  let actualBoxId: string | null = null; // To store the obtained box ID
  let browser: BoxBrowserManager | null = null; // Declare browser outside try, use any for now
  let context: BrowserContext | null = null; // Declare context outside try, use any for now
  let page: BrowserPage | null = null; // Declare page outside try, use any for now

  try {
    // Get or create box
    const boxResult = await gbox.boxes.getOrCreateBox({
      boxId,
      image: config.images.playwright,
      sessionId,
      signal,
      waitTimeoutSeconds: 60,
    });

    // Check if image is being pulled
    if (boxResult.imagePullStatus?.inProgress) {
      logger.info(`Image ${boxResult.imagePullStatus.imageName} is being pulled: ${boxResult.imagePullStatus.message}`);
      return {
        content: [
          {
            type: "text" as const,
            text: `Image is being pulled: ${boxResult.imagePullStatus.message}`,
          },
        ],
      };
    }

    // Ensure we have a valid boxId
    if (!boxResult.boxId) {
      logger.error("Failed to get or create box");
      return {
        content: [
          {
            type: "error" as const,
            text: "Failed to get or create box",
          },
        ],
      };
    }

    actualBoxId = boxResult.boxId;

    logger.info(
      `Attempting to view URL: ${url} as ${as} in box ${actualBoxId} ${
        sessionId ? `for session: ${sessionId}` : ""
      }`
    );

    // Instantiate BrowserContextOperations with obtained boxId
    browser = await gbox.boxes.initBrowser(actualBoxId);

    // 1. Create a new browser context
    context = await browser.createContext({}, signal); // Assign to the outer context variable
    if (!context) {
      logger.error("Failed to create browser context");
      return { content: [{ type: "error", text: "Failed to create browser context" }] };
    }
    contextId = context.id;
    logger.debug(`Created browser context: ${contextId}`);

    // 2. Create a new page and navigate to the URL
    const pageParams = {
      url,
      wait_until: "load" as const, // Use const assertion
    };
    page = await context.createPage(pageParams, signal); // Assign to the outer page variable
    pageId = page.id;
    logger.debug(`Created page ${pageId} and navigated to URL: ${url}`);

    let resultText: string | undefined; // Make optional as it's not always set
    switch (as) {
      case "html": {
        const result = await page.getContent();
        logger.info(
          `Fetched HTML content for ${url}. Length: ${
            result.content?.length ?? 0
          }`
        );
        resultText = result.content ?? "";
        break;
      }
      case "markdown": {
        const result = await page.getContent("markdown");
        logger.info(
          `Fetched Markdown content for ${url}. Length: ${
            result.content?.length ?? 0
          }`
        );
        resultText = result.content ?? "";
        break;
      }
      case "screenshot": {
        const screenshotParams: VisionScreenshotParams = {
          outputFormat: "base64" as const,
        };
        
        const result: VisionScreenshotResult | BrowserErrorResult = await page.screenshot(screenshotParams);
        if (result.success === true && result.base64_content) {
          logger.info(
            `Took screenshot for ${url}, returning base64 data (length: ${result.base64_content.length}).`
          );
          let mimeType = "image/png";
          return {
            content: [
              {
                type: "image" as const,
                data: result.base64_content,
                mimeType: mimeType, // Using assumed PNG mime type
              },
            ],
          };
          } else if (result.success === true && !result.base64_content) {
          const errorMsg = `Screenshot succeeded but base64 content is missing.`;
          logger.error(errorMsg);
          return { content: [{ type: "error", text: errorMsg }] };
        } else if (isBrowserErrorResult(result)) {
          const errorMsg = `Screenshot failed: ${result.error}`;
          logger.error(`Failed to take screenshot for ${url}: ${result.error}`);
          return { content: [{ type: "error", text: errorMsg }] };
        } else {
          const errorMsg = `Screenshot failed with unexpected result format.`;
          logger.error(
            `Unexpected screenshot result format for ${url}: ${JSON.stringify(
              result
            )}`
          );
          return { content: [{ type: "error", text: errorMsg }] };
        }
        // No resultText is set here, return happened above or in error cases
        break; // Break is technically redundant due to returns, but good practice
      }
      default: {
        const exhaustiveCheck: never = as;
        const errorMsg = `Internal error: Unhandled format '${exhaustiveCheck}'`;
        logger.error(`Unhandled 'as' value: ${exhaustiveCheck}`);
        return { content: [{ type: "error", text: errorMsg }] };
      }
    }

    // Return success content for text-based formats if resultText was set
    if (resultText !== undefined) {
      return { content: [{ type: "text", text: resultText }] };
    } else {
      // Should ideally not be reached if all paths return or throw
      const errorMsg = "Internal error: Reached end of switch without result.";
      logger.error(errorMsg);
      return { content: [{ type: "error", text: errorMsg }] };
    }
  } catch (error: any) {
    const errorMsg = `Operation failed: ${error.message || String(error)}`;
    logger.error(
      `Error viewing URL ${url} as ${as}: ${error.message || error}`
    );
    // Return error content
    return { content: [{ type: "error", text: errorMsg }] };
  } finally {
    // 4. Ensure cleanup: Close page and context
    // Note: Signal is not typically passed to cleanup operations
    if (pageId) { // Check if page was assigned
      try {
        logger.debug(`Closing page ${pageId}...`);
        await context?.closePage(pageId);
      } catch (closePageError: any) {
        logger.warn(
          `Failed to close page ${pageId}: ${
            closePageError.message || closePageError
          }`
        );
      }
    }
    if (contextId) { // Check if context was assigned
      try {
        logger.debug(`Closing context ${contextId}...`);
        await browser?.closeContext(contextId);
      } catch (closeContextError: any) {
        logger.warn(
          `Failed to close context ${contextId}: ${
            closeContextError.message || closeContextError
          }`
        );
      }
    }
  }
}

// Export the wrapped function
export const handleViewUrlAs = withLogging(_handleViewUrlAs);

export function isBrowserErrorResult(obj: any): obj is BrowserErrorResult {
  return obj && obj.success === false && typeof obj.error === "string";
}

export interface BrowserErrorResult {
  success: false;
  error: string;
}
