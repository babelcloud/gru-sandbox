import type { Client } from "./client";
import type {
  CreateContextParams,
  CreateContextResult,
  CreatePageParams,
  CreatePageResult,
  GetPageParams,
  GetPageResult,
  ListPagesResult,
  VisionScreenshotParams,
  VisionScreenshotResult,
  BrowserErrorResult,
} from "./types";

/**
 * Operations related to a specific browser page within a context.
 */
export class PageOperations {
  constructor(
    private apiClient: Client,
    private boxId: string,
    private contextId: string,
    private pageId: string
  ) {}

  /**
   * Get details for this specific page.
   * @param params Options for retrieving content.
   * @returns Page details, potentially including content.
   */
  async get(params?: GetPageParams): Promise<GetPageResult> {
    const queryParams = new URLSearchParams();
    if (params?.withContent !== undefined) {
      queryParams.set("withContent", String(params.withContent));
    }
    if (params?.contentType) {
      queryParams.set("contentType", params.contentType);
    }
    const queryString = queryParams.toString();
    const url = `/boxes/${this.boxId}/browser-contexts/${
      this.contextId
    }/pages/${this.pageId}${queryString ? `?${queryString}` : ""}`;
    const response = await this.apiClient.get(url);
    return response.json();
  }

  /**
   * Close this specific page.
   */
  async close(): Promise<void> {
    const url = `/boxes/${this.boxId}/browser-contexts/${this.contextId}/pages/${this.pageId}`;
    await this.apiClient.delete(url);
  }

  /**
   * Take a screenshot of the current page.
   * @param params Screenshot options.
   * @returns Result containing the path to the saved screenshot or an error.
   */
  async screenshot(
    params: VisionScreenshotParams = {}
  ): Promise<VisionScreenshotResult | BrowserErrorResult> {
    const url = `/boxes/${this.boxId}/browser-contexts/${this.contextId}/pages/${this.pageId}/actions/vision-screenshot`;
    const response = await this.apiClient.post(url, {
      body: JSON.stringify(params),
    });
    return response.json();
  }

  // TODO: Add methods for Vision Actions (click, type, screenshot, etc.)
  // Example:
  // async click(params: VisionClickParams): Promise<VisionClickResult | BrowserErrorResult> {
  //   const url = `/boxes/${this.boxId}/browser-contexts/${this.contextId}/pages/${this.pageId}/actions/vision-click`;
  //   return this.apiClient.request<VisionClickResult | BrowserErrorResult>('POST', url, params);
  // }
}

/**
 * Operations related to pages within a specific browser context.
 */
export class ContextPageOperations {
  constructor(
    private apiClient: Client,
    private boxId: string,
    private contextId: string
  ) {}

  /**
   * Create a new page within this context and navigate it.
   * @param params Page creation parameters (URL, etc.).
   * @returns Result of the page creation.
   */
  async create(params: CreatePageParams): Promise<CreatePageResult> {
    const url = `/boxes/${this.boxId}/browser-contexts/${this.contextId}/pages`;
    const response = await this.apiClient.post(url, {
      body: JSON.stringify(params),
    });
    return response.json();
  }

  /**
   * List all pages within this context.
   * @returns List of page IDs.
   */
  async list(): Promise<ListPagesResult> {
    const url = `/boxes/${this.boxId}/browser-contexts/${this.contextId}/pages`;
    const response = await this.apiClient.get(url);
    return response.json();
  }

  /**
   * Get operations for a specific page within this context.
   * @param pageId The ID of the page.
   * @returns PageOperations instance.
   */
  page(pageId: string): PageOperations {
    return new PageOperations(
      this.apiClient,
      this.boxId,
      this.contextId,
      pageId
    );
  }
}

/**
 * Operations related to browser contexts within a specific box.
 */
export class BrowserContextOperations {
  constructor(private apiClient: Client, private boxId: string) {}

  /**
   * Create a new browser context within this box.
   * @param params Context creation parameters.
   * @returns Result of the context creation.
   */
  async create(params?: CreateContextParams): Promise<CreateContextResult> {
    const url = `/boxes/${this.boxId}/browser-contexts`;
    const response = await this.apiClient.post(url, {
      body: JSON.stringify(params ?? {}),
    });

    // Check if the request was successful
    if (!response.ok) {
      // Attempt to read the response body as text for more detailed error logging
      let errorBody = await response
        .text()
        .catch(() => "Failed to read error body");
      throw new Error(
        `Failed to create browser context: ${response.status} ${response.statusText}. Body: ${errorBody}`
      );
    }

    return response.json();
  }

  /**
   * Close a specific browser context within this box.
   * @param contextId The ID of the context to close.
   */
  async close(contextId: string): Promise<void> {
    const url = `/boxes/${this.boxId}/browser-contexts/${contextId}`;
    await this.apiClient.delete(url);
  }

  /**
   * Get operations for pages within a specific context.
   * @param contextId The ID of the browser context.
   * @returns ContextPageOperations instance.
   */
  pages(contextId: string): ContextPageOperations {
    return new ContextPageOperations(this.apiClient, this.boxId, contextId);
  }
}
