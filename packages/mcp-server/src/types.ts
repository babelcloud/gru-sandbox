import type { Server } from "@modelcontextprotocol/sdk/server/index.js";

/**
 * Represents the result of running a command in a box
 */
export interface BoxRunResult {
  boxId: string;
  exitCode: number;
  stdout: string;
  stderr: string;
}

/**
 * Options for running commands in a box
 */
export interface RunOptions {
  boxId?: string;
  sessionId?: string;
  signal?: AbortSignal;
}

/**
 * Options for creating a box
 */
export interface CreateBoxOptions {
  boxId?: string;
  image: string;
  sessionId?: string;
  signal?: AbortSignal;
}

/**
 * Represents a box in the system
 */
export interface Box {
  id: string;
  status: string;
  image: string;
  labels?: Record<string, string>;
}



export type LogFn = Server["sendLoggingMessage"];
