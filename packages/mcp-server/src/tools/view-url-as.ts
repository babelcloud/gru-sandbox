import { z } from "zod";
import { GBox } from "../sdk"; // Assuming GBox provides client and boxId
import { Client } from "../sdk/client"; // Import Client type
import {
  BrowserContextOperations,
  ContextPageOperations,
  PageOperations,
} from "../sdk/browser"; // Import operation classes
import type {
  Logger,
  VisionScreenshotParams, // Import the specific type
  VisionScreenshotResult, // Import the specific type
  BrowserErrorResult, // Added missing import
} from "../sdk/types";
import { isBrowserErrorResult } from "../sdk/types"; // Import type guard if needed
import { withLogging } from "../utils.js"; // Added import
import { config } from "../config.js"; // Added import

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
  const gbox = new GBox({
    apiUrl: config.apiServer.url,
    logger,
  });

  // Note: Assuming gbox.client is the correct way to get the Client instance
  const client = gbox.client; // Get client from gbox instance

  if (!client) {
    // Check if client is available
    const errorMsg = "GBox client is missing for browser operations.";
    logger.error(errorMsg);
    return { content: [{ type: "error", text: errorMsg }] };
  }

  let contextOps: BrowserContextOperations | null = null;
  let pageOps: PageOperations | null = null;
  let contextId: string | null = null;
  let pageId: string | null = null;
  let actualBoxId: string | null = null; // To store the obtained box ID

  try {
    // Get or create box
    actualBoxId = await gbox.box.getOrCreateBox({
      boxId,
      image: config.images.playwright,
      sessionId,
      signal,
      waitForReady: true, // Wait for the box healthcheck to pass
      waitForReadyTimeoutSeconds: 60, // Timeout after 60 seconds
    });

    logger.info(
      `Attempting to view URL: ${url} as ${as} in box ${actualBoxId} ${
        sessionId ? `for session: ${sessionId}` : ""
      }`
    );

    // Instantiate BrowserContextOperations with obtained boxId
    contextOps = new BrowserContextOperations(client, actualBoxId);

    // 1. Create a new browser context
    const contextResult = await contextOps.create(/* { signal } */);
    contextId = contextResult.context_id;
    logger.debug(`Created browser context: ${contextId}`);

    // Instantiate ContextPageOperations with obtained boxId
    const contextPageOps = new ContextPageOperations(
      client,
      actualBoxId,
      contextId
    );

    // 2. Create a new page and navigate to the URL
    const pageParams = {
      url,
      wait_until: "load" as const, // Use const assertion
    };
    const pageResult = await contextPageOps.create(pageParams);
    pageId = pageResult.page_id;
    logger.debug(`Created page ${pageId} and navigated to URL: ${url}`);

    // Instantiate PageOperations with obtained boxId
    pageOps = new PageOperations(client, actualBoxId, contextId, pageId);

    // 3. Perform the requested action
    let resultText: string | undefined; // Make optional as it's not always set
    switch (as) {
      case "html": {
        const getParams = {
          withContent: true,
          contentType: "html" as const, // Use const assertion
        };
        const result = await pageOps.get(getParams);
        logger.info(
          `Fetched HTML content for ${url}. Length: ${
            result.content?.length ?? 0
          }`
        );
        resultText = result.content ?? "";
        break;
      }
      case "markdown": {
        const getParams = {
          withContent: true,
          contentType: "markdown" as const, // Use const assertion
        };
        const result = await pageOps.get(getParams);
        logger.info(
          `Fetched Markdown content for ${url}. Length: ${
            result.content?.length ?? 0
          }`
        );
        resultText = result.content ?? "";
        break;
      }
      case "screenshot": {
        // Specify base64 output format and pass type from input params
        const screenshotParams: VisionScreenshotParams = {
          output_format: "base64" as const,
          // Determine type based on the 'as' parameter logic (though screenshot always returns png/jpeg)
          // The backend decides the actual format, but we can hint based on future extensions
          // For now, the backend defaults to png if type is not jpeg.
          // type: as === 'screenshot' ? undefined : as, // This logic is flawed
        };
        // Potentially refine type based on future params, for now API handles default

        const result: VisionScreenshotResult | BrowserErrorResult =
          await pageOps.screenshot(screenshotParams);

        if (result.success === true && result.base64_content) {
          logger.info(
            `Took screenshot for ${url}, returning base64 data (length: ${result.base64_content.length}).`
          );

          // Determine MIME type - NEEDS IMPROVEMENT
          // Ideally, the backend API should return the actual mime type.
          // For now, we assume PNG unless the backend explicitly signals JPEG (which it doesn't directly).
          // We *cannot* reliably infer the type from the base64 string itself without more complex logic.
          // Defaulting to png as per current backend behavior.
          let mimeType = "image/png";
          // if (screenshotParams.type === "jpeg") { // Cannot reliably check screenshotParams.type here
          //    mimeType = "image/jpeg";
          // }

          // Return the structured image content directly
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
    if (pageOps) {
      try {
        logger.debug(`Closing page ${pageId}...`);
        await pageOps.close();
      } catch (closePageError: any) {
        logger.warn(
          `Failed to close page ${pageId}: ${
            closePageError.message || closePageError
          }`
        );
      }
    }
    if (contextOps && contextId) {
      try {
        logger.debug(`Closing context ${contextId}...`);
        await contextOps.close(contextId);
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
