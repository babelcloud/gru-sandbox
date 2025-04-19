// Logger interface that can be implemented by different logging systems
export interface Logger {
  debug(message: string, ...args: any[]): void;
  info(message: string, ...args: any[]): void;
  warn(message: string, ...args: any[]): void;
  error(message: string, ...args: any[]): void;
}

// SDK Configuration
export interface GBoxConfig {
  apiUrl: string;
  logger?: Logger;
}

// Common types shared across the SDK
export interface Box {
  id: string;
  status: string;
  image: string;
  labels?: Record<string, string>;
}

export interface BoxRunResult {
  box: Box;
  exitCode: number;
  stdout: string;
  stderr: string;
}

export interface CreateBoxOptions {
  boxId?: string;
  image: string;
  sessionId?: string;
  signal?: AbortSignal;
  waitForReady?: boolean;
  waitForReadyTimeoutSeconds?: number;
}

export interface RunOptions {
  sessionId?: string;
  signal?: AbortSignal;
}

// File related types
export interface FileStat {
  name: string;
  path: string;
  size: number;
  mode: string;
  modTime: string;
  type: string;
  mime: string;
}

export interface FileMetadataResponse {
  fileStat: FileStat;
  mimeType: string;
  contentLength: number;
}

export interface FileShareResponse {
  success: boolean;
  message: string;
  fileList: FileStat[];
}

// Constants
export const FILE_SIZE_LIMITS = {
  TEXT: 1024 * 1024, // 1MB for text files
  BINARY: 5 * 1024 * 1024, // 5MB for binary files (images, audio)
} as const;

// --- Browser Context Types ---

export interface CreateContextParams {
  // Add options if defined in Go model (e.g., playwright options)
  // Example: userAgent?: string;
}

export interface CreateContextResult {
  context_id: string;
  // Add other fields if defined in Go model
}

// --- Browser Page Types ---

// Map playwright WaitUntilState (string enums in Go likely)
export type WaitUntilState =
  | "load"
  | "domcontentloaded"
  | "networkidle"
  | "commit";

export interface CreatePageParams {
  url: string;
  wait_until?: WaitUntilState;
  timeout?: number; // milliseconds
}

export interface CreatePageResult {
  page_id: string;
  url: string;
  title: string;
}

export interface ListPagesResult {
  page_ids: string[];
}

export interface GetPageParams {
  withContent?: boolean;
  contentType?: "html" | "markdown";
}

export interface GetPageResult {
  page_id: string;
  url: string;
  title: string;
  content?: string;
  contentType?: "text/html" | "text/markdown"; // MIME types
}

// --- Browser Action Error Types ---

// Using a generic error result type for actions for now
// Matches Go's model.VisionErrorResult structure
export interface BrowserErrorResult {
  success: false;
  error: string;
}

// Type guard to check for browser errors
export function isBrowserErrorResult(obj: any): obj is BrowserErrorResult {
  return obj && obj.success === false && typeof obj.error === "string";
}

// --- Vision Action Specific Types ---

// Corresponds to Go's model.Rect
export interface Rect {
  x: number;
  y: number;
  width: number;
  height: number;
}

// Corresponds to Go's model.VisionScreenshotParams
export interface VisionScreenshotParams {
  // path?: string; // REMOVED: No longer specified by client
  type?: "png" | "jpeg";
  quality?: number; // 0-100, only for jpeg
  fullPage?: boolean;
  clip?: Rect;
  omitBackground?: boolean;
  timeout?: number; // milliseconds
  scale?: "css" | "device";
  animations?: "disabled" | "allow";
  caret?: "hide" | "initial";
  output_format?: "url" | "base64"; // ADDED: Specify output format
  // Add other options if needed based on Go struct
}

// Corresponds to Go's model.VisionScreenshotResult
export interface VisionScreenshotResult {
  success: true;
  // savedPath: string; // REMOVED: Replaced by URL or base64 content
  url?: string; // ADDED: URL if output_format is "url"
  base64_content?: string; // ADDED: Base64 content if output_format is "base64"
}

// Consider adding other Vision Action types here later if needed
// e.g., VisionClickParams, VisionClickResult, etc.
