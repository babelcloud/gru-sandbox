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
  exitCode: number;
  stdout: string;
  stderr: string;
}

export interface CreateBoxOptions {
  boxId?: string;
  image: string;
  sessionId?: string;
  signal?: AbortSignal;
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
