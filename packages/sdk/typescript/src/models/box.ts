import { BoxApi } from '../api/box.api.ts';
import { BrowserApi } from '../api/browser.api.ts'; // Import BrowserApi
// Use import type for interfaces/types
import type {
  BoxData,
  BoxReclaimResponse,
  BoxRunResponse,
  BoxExtractArchiveResponse,
  BoxRunOptions, // Import the new options type
  BoxExecProcess,
  BoxExecOptions, // Pass through new options type
} from '../types/box.ts';
import { NotFoundError } from '../errors.ts';
// --- Node.js imports ---
import * as fs from 'node:fs';
import * as path from 'node:path';
import * as tar from 'tar'; // Use the installed tar package
import type { ReadEntry } from 'tar'; // Import specific type
import { Readable } from 'node:stream'; // For converting ArrayBuffer to stream

import { BoxBrowserManager } from '../managers/browser.manager.ts'; // Import the new manager
import { logger } from '../logger.ts';

/**
 * Represents a GBox Box instance.
 *
 * Provides methods to interact with a specific Box via getters and instance methods.
 * Attributes are stored in the `attrs` property.
 */
export class Box {
  // Store the core Box data
  public attrs: BoxData;
  // Keep references to the API layers for instance methods
  private readonly boxApi: BoxApi;
  private readonly browserApi: BrowserApi; // Add browserApi property

  // Constructor now takes BoxData, BoxApi, and BrowserApi (logger removed)
  constructor(boxData: BoxData, boxApi: BoxApi, browserApi: BrowserApi) {
    this.attrs = boxData; // Store the initial attributes
    this.boxApi = boxApi;
    this.browserApi = browserApi; // Store browserApi
  }

  // --- Getters for accessing attributes ---

  get id(): string {
    return this.attrs.id;
  }

  get status(): string {
    return this.attrs.status;
  }

  get image(): string {
    return this.attrs.image;
  }

  get labels(): Record<string, string> | undefined {
    return this.attrs.labels;
  }

  // --- Instance Methods ---

  /**
   * Updates the Box instance's attributes by fetching the latest data from the API.
   * @param signal An optional AbortSignal to cancel the operation.
   */
  async reload(signal?: AbortSignal): Promise<void> {
    try {
      // Use the getter for id
      const updatedData = await this.boxApi.getDetails(this.id, signal);
      this.attrs = updatedData; // Update the entire attrs object
    } catch (error) {
      // Handle cases where the box might no longer exist
      if (error instanceof NotFoundError) {
        // Optionally update status or throw a more specific error
        this.attrs.status = 'deleted'; // Example: Mark as deleted
        logger.warn(
          `[Box ID: ${this.id}] Failed to reload, Box not found. Marked as deleted.`,
          error
        );
      } else {
        // Re-throw other errors
        throw error;
      }
    }
  }

  /**
   * Starts the Box.
   * @param signal An optional AbortSignal to cancel the operation.
   */
  async start(
    signal?: AbortSignal
  ): Promise<{ success: boolean; message: string }> {
    // Use the getter for id
    const result = await this.boxApi.start(this.id, signal);
    await this.reload(signal); // Update box status after action
    return result;
  }

  /**
   * Stops the Box.
   * @param signal An optional AbortSignal to cancel the operation.
   */
  async stop(
    signal?: AbortSignal
  ): Promise<{ success: boolean; message: string }> {
    // Use the getter for id
    const result = await this.boxApi.stop(this.id, signal);
    await this.reload(signal); // Update box status after action
    return result;
  }

  /**
   * Deletes the Box.
   * @param signal An optional AbortSignal to cancel the operation.
   */
  async delete(
    force: boolean = false,
    signal?: AbortSignal
  ): Promise<{ message: string }> {
    // Note: After deletion, this Box instance becomes stale.
    // Use the getter for id
    const result = await this.boxApi.deleteBox(this.id, force, signal);
    return result;
  }

  /**
   * Runs a command inside the Box.
   * @param signal An optional AbortSignal to cancel the operation.
   */
  async run(
    command: string[],
    options?: BoxRunOptions,
    signal?: AbortSignal
  ): Promise<BoxRunResponse> {
    const response = await this.boxApi.run(this.id, command, options, signal);
    // Handle cases where API might return 200 OK without an exitCode.
    // Default to -1 (unknown/not provided) if missing, mimicking Python SDK.
    if (response.exitCode === undefined || response.exitCode === null) {
      logger.debug(`[Box ID: ${this.id}] Run command response missing exitCode, defaulting to -1.`);
      response.exitCode = -1;
    }
    logger.info(`[Box ID: ${this.id}] Run command completed. Exit code: ${response.exitCode}`);
    return response;
  }

  /**
   * Executes a command in the Box via WebSocket, waits for completion,
   * and returns streams for stdout/stderr and a promise for the exit code.
   *
   * @param cmd The command and its arguments.
   * @param options Optional settings like tty mode, working directory, and abort signal.
   * @returns A Promise resolving to an object containing `stdout` and `stderr` Readable streams, and a Promise `exitCode` which resolves with the command's exit code.
   */
  async exec(
    cmd: string[], 
    options?: BoxExecOptions
  ): Promise<BoxExecProcess> {
    // Pass cmd and options directly to the api method
    const result = await this.boxApi.exec(this.id, cmd, options);
    return result;
  }

  /**
   * Reclaims resources associated with this specific Box.
   * @param signal An optional AbortSignal to cancel the operation.
   */
  async reclaim(
    force: boolean = false,
    signal?: AbortSignal
  ): Promise<BoxReclaimResponse> {
    // Use the getter for id
    const result = await this.boxApi.reclaim(this.id, force, signal);
    await this.reload(signal); // Update box status after action
    return result;
  }

  /**
   * Copies files/directories from the host to this Box (using archives).
   *
   * @param sourcePath The local path to the file or directory to copy.
   * @param targetPath The destination directory path inside the Box.
   * @param signal An optional AbortSignal to cancel the operation.
   */
  async copyTo(
    sourcePath: string,
    targetPath: string,
    signal?: AbortSignal
  ): Promise<BoxExtractArchiveResponse> {
    logger.debug(`[Box ID: ${this.id}] Attempting copyTo: ${sourcePath} -> ${targetPath}`);
    // Check if source exists using Node.js fs
    try {
      await fs.promises.stat(sourcePath);
      logger.debug(`[Box ID: ${this.id}] copyTo: Source path ${sourcePath} exists.`);
    } catch (error: any) {
      // Catch specific error type if possible
      if (error.code === 'ENOENT') {
        logger.error(`[Box ID: ${this.id}] copyTo: Local source path not found: ${sourcePath}`);
        // Node.js error code for Not Found
        throw new Error(`Local source path not found: ${sourcePath}`);
      }
      logger.error(`[Box ID: ${this.id}] copyTo: Error stating source path ${sourcePath}:`, error);
      throw error; // Re-throw other errors
    }

    logger.debug(`[Box ID: ${this.id}] copyTo: Creating tar archive in memory for ${sourcePath}.`);
    // Create tar archive in memory using the 'tar' package
    // 'tar' creates a stream, we need to collect it into a buffer
    const tarStream = tar.c(
      {
        gzip: false, // Box API expects uncompressed tar
        cwd: path.dirname(sourcePath), // Work relative to the source file's directory
        // prefix: path.basename(sourcePath) // Removing prefix, let tar handle the structure
      },
      [path.basename(sourcePath)] // Add just the basename relative to cwd
    );

    const chunks: Buffer[] = [];
    for await (const chunk of tarStream) {
      chunks.push(chunk instanceof Buffer ? chunk : Buffer.from(chunk));
    }
    const archiveDataBuffer = Buffer.concat(chunks);
    logger.debug(`[Box ID: ${this.id}] copyTo: Tar archive created in memory, size: ${archiveDataBuffer.length} bytes.`);
    // Convert Node.js Buffer to ArrayBuffer for the API call
    const archiveData = archiveDataBuffer.buffer.slice(
      archiveDataBuffer.byteOffset,
      archiveDataBuffer.byteOffset + archiveDataBuffer.byteLength
    );

    // Use the getter for id and call the API
    const result = await this.boxApi.extractArchive(this.id, targetPath, archiveData, signal);
    return result;
  }

  /**
   * Copies files/directories from this Box to the host.
   * If localPath is provided, extracts the archive content to that path.
   * Otherwise, returns the raw tar archive data as ArrayBuffer.
   *
   * @param sourcePath The path to the file or directory inside the Box.
   * @param localPath Optional. The local path to extract the content to.
   * @param signal An optional AbortSignal to cancel the operation.
   */
  async copyFrom(
    sourcePath: string,
    localPath?: string,
    signal?: AbortSignal
  ): Promise<ArrayBuffer | void> {
    // Use the getter for id to get the archive data
    const archiveData: ArrayBuffer = await this.boxApi.getArchive(
      this.id,
      sourcePath,
      signal
    );
    
    if (localPath) {
      logger.debug(`[Box ID: ${this.id}] copyFrom: Processing localPath: ${localPath}`);
      // Determine if the target is a directory or a file path, and ensure the base directory exists
      let extractBaseDir = localPath;
      let isTargetFile = false;

      try {
        // Check if localPath exists
        const stats = await fs.promises.stat(localPath);
        if (!stats.isDirectory()) {
          // It exists but is not a directory, treat as file target
          extractBaseDir = path.dirname(localPath);
          isTargetFile = true;
          logger.debug(`[Box ID: ${this.id}] copyFrom: localPath ${localPath} exists and is a file. Extract base dir: ${extractBaseDir}`);
        }
        // If it's an existing directory, extractBaseDir remains localPath
      } catch (e: any) {
        if (e.code === 'ENOENT') {
          // Path doesn't exist. Assume it's a file if it has an extension or doesn't end with sep.
          if (path.extname(localPath) || !localPath.endsWith(path.sep)) {
            // Intended as a file path
            extractBaseDir = path.dirname(localPath);
            isTargetFile = true;
            logger.debug(`[Box ID: ${this.id}] copyFrom: localPath ${localPath} does not exist, treated as file. Extract base dir: ${extractBaseDir}`);
          } else {
            // Intended as a directory path, extractBaseDir remains localPath
            isTargetFile = false;
            logger.debug(`[Box ID: ${this.id}] copyFrom: localPath ${localPath} does not exist, treated as directory.`);
          }
          // Ensure the base directory exists (for both file and dir targets)
          logger.debug(`[Box ID: ${this.id}] copyFrom: Ensuring directory ${extractBaseDir} exists.`);
          await fs.promises.mkdir(extractBaseDir, { recursive: true });
        } else {
          logger.error(`[Box ID: ${this.id}] copyFrom: Error stating localPath ${localPath}:`, e);
          throw e; // Re-throw other stat errors
        }
      }

      // --- Handle file download directly, directory download via tar ---
      if (isTargetFile) {
        logger.debug(`[Box ID: ${this.id}] copyFrom: Extracting single file to ${localPath}.`);
        // Use tar parser to extract the single file content
        const buffer = Buffer.from(archiveData);
        const readableStream = Readable.from(buffer);

        await new Promise<void>((resolve, reject) => {
          const parser = new tar.Parser();
          let fileStreamOpened = false;

          parser.on('entry', (entry: ReadEntry) => {
            logger.debug(`[Box ID: ${this.id}] copyFrom: Tar entry: ${entry.path}, type: ${entry.type}`);
            // Assuming the first file entry is the one we want
            if (entry.type === 'File' && !fileStreamOpened) {
              fileStreamOpened = true; // Process only the first file entry
              logger.debug(`[Box ID: ${this.id}] copyFrom: Piping entry ${entry.path} to ${localPath}.`);
              const writeStream = fs.createWriteStream(localPath);

              entry
                .pipe(writeStream)
                .on('finish', () => {
                  logger.debug(`[Box ID: ${this.id}] copyFrom: Finished writing file ${localPath}.`);
                  // Need to ensure parser also finishes if needed, but writeStream finish is key
                  resolve();
                })
                .on('error', (writeErr: Error) => {
                  logger.error(`[Box ID: ${this.id}] copyFrom: Failed to write to ${localPath}:`, writeErr);
                  reject(
                    new Error(
                      `Failed to write to ${localPath}: ${writeErr.message}`
                    )
                  );
                });
            } else {
              // Drain other entries (like directories if API includes them unexpectedly)
              logger.debug(`[Box ID: ${this.id}] copyFrom: Draining tar entry: ${entry.path}`);
              entry.resume();
            }
          });

          parser.on('end', () => {
            logger.debug(`[Box ID: ${this.id}] copyFrom: Tar parser finished.`);
            // Resolve if writeStream hasn't already (e.g., empty tar?)
            // Or potentially reject if no file entry was found?
            if (!fileStreamOpened) {
              logger.warn(`[Box ID: ${this.id}] copyFrom: Tar parser finished but no file entry was processed for ${localPath}.`);
              // It's possible the archive was empty or didn't contain a file entry.
              // Or maybe the API returned a tar for a directory even when a file path was requested?
              // Resolve for now, but maybe should reject if no file was written?
              // Consider if API guarantees a single file entry in this case.
              resolve();
            }
          });

          parser.on('error', (parseErr: Error) => {
            logger.error(`[Box ID: ${this.id}] copyFrom: Failed to parse tar archive:`, parseErr);
            reject(
              new Error(`Failed to parse tar archive: ${parseErr.message}`)
            );
          });

          readableStream.pipe(parser);
        });
      } else {
        logger.debug(`[Box ID: ${this.id}] copyFrom: Extracting archive to directory ${extractBaseDir}.`);
        // Extract the archive using the 'tar' package for directories
        const buffer = Buffer.from(archiveData);
        const readableStream = Readable.from(buffer);
        await new Promise<void>((resolve, reject) => {
          const extractor = tar.x({
            cwd: extractBaseDir, // Extract into the determined base directory
            strip: 0, // Don't strip components, handle potential nesting manually if needed
          });

          extractor.on('finish', () => {
            logger.debug(`[Box ID: ${this.id}] copyFrom: Finished extracting archive to ${extractBaseDir}.`);
            resolve();
          });
          extractor.on('error', (extractErr: Error) => {
            logger.error(`[Box ID: ${this.id}] copyFrom: Failed to extract archive to ${extractBaseDir}:`, extractErr);
            reject(extractErr);
          });

          readableStream.pipe(extractor);
        });
      }
      // --- End handling ---
      logger.info(`[Box ID: ${this.id}] copyFrom: Successfully copied ${sourcePath} to ${localPath}.`);
      return; // Return void when extracting locally
    } else {
      logger.debug(`[Box ID: ${this.id}] copyFrom: No localPath provided, returning raw archive data.`);
      // Return raw ArrayBuffer if no local path is provided
      return archiveData;
    }
  }

  /**
   * Gets metadata about a file or directory inside the Box.
   * @param signal An optional AbortSignal to cancel the operation.
   */
  async stat(
    path: string,
    signal?: AbortSignal
  ): Promise<Record<string, string>> {
    // Use the getter for id
    const result = await this.boxApi.headArchive(this.id, path, signal);
    return result;
  }

  /**
   * Gets a manager instance for handling browser contexts within this Box.
   * @returns A BoxBrowserManager instance scoped to this Box.
   */
  initBrowser(): BoxBrowserManager {
    const manager = new BoxBrowserManager(this.id, this.browserApi);
    return manager;
  }

  // Potentially add listBrowserContexts, getBrowserContext in the future if API supports
}
