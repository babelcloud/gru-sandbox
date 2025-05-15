import { Client } from './http-client.ts';
import type {
  BoxCreateOptions,
  BoxCreateResponse,
  BoxData,
  BoxGetResponse,
  BoxListFilters,
  BoxListResponse,
  BoxesDeleteResponse,
  BoxReclaimResponse,
  BoxRunResponse,
  BoxExtractArchiveResponse,
  BoxRunOptions,
  StreamType,
  ExitMessage,
  BoxExecProcess,
  BoxExecOptions,
} from '../types/box.ts';
import { StreamTypeStdout, StreamTypeStderr } from '../types/box.ts';
import { WebSocketClient } from './ws-client.ts';
import { logger } from '../logger.ts';
import fs from 'fs';
const API_PREFIX = '/api/v1/boxes';

export class ImagePullInProgressError extends Error {
  imageName: string;
  
  constructor(message: string, imageName: string) {
    super(message);
    this.name = 'ImagePullInProgressError';
    this.imageName = imageName;
  }
}

export class BoxApi extends Client {
  /**
   * List boxes with optional filters.
   * GET /api/v1/boxes
   */
  async list(
    filters?: BoxListFilters,
    signal?: AbortSignal
  ): Promise<BoxListResponse> {
    let params: Record<string, string | string[]> = {};
    if (filters) {
      const filterParams: string[] = [];
      for (const [key, value] of Object.entries(filters)) {
        if (Array.isArray(value)) {
          value.forEach((item) => filterParams.push(`${key}=${item}`));
        } else if (value !== undefined) {
          filterParams.push(`${key}=${value}`);
        }
      }
      if (filterParams.length > 0) {
        params['filter'] = filterParams;
      }
    }
    const response = await super.get<BoxListResponse>(
      API_PREFIX,
      params,
      undefined,
      signal
    );
    response.boxes = response.boxes.map((box) => this.mapLabels(box));
    return response;
  }

  /**
   * Get details of a specific box.
   * GET /api/v1/boxes/{id}
   */
  async getDetails(
    boxId: string,
    signal?: AbortSignal
  ): Promise<BoxGetResponse> {
    const response = await super.get<BoxGetResponse>(
      `${API_PREFIX}/${boxId}`,
      undefined,
      undefined,
      signal
    );
    return this.mapLabels(response);
  }

  /**
   * Create a new box.
   * POST /api/v1/boxes
   * 
   * @param options - Box creation options
   * @param options.timeout - Duration string (e.g., '30s', '1m') for image pull timeout. 
   *                          If specified and the image doesn't exist locally, the API will:
   *                          1. Start an async image pull
   *                          2. Return a response with imagePullStatus information instead of throwing
   *                          3. The client should monitor or retry creation after waiting
   * @param signal - Optional AbortSignal to cancel the request
   * @returns The created box data, or a response with imagePullStatus if image is being pulled
   * 
   * @example
   * ```typescript
   * // Create a box with timeout
   * const response = await client.boxes.create({
   *   image: "my-image:latest",
   *   timeout: "30s" // Wait up to 30 seconds for image pull
   * });
   * 
   * if (response.imagePullStatus?.inProgress) {
   *   console.log(`Image ${response.imagePullStatus.imageName} is being pulled: ${response.imagePullStatus.message}`);
   *   // Implement retry logic or notify user to try again
   * } else {
   *   console.log("Box created successfully:", response.id);
   * }
   * ```
   */
  async create(
    options: BoxCreateOptions,
    signal?: AbortSignal
  ): Promise<BoxCreateResponse> {
    const apiOptions: Record<string, any> = { ...options };
    if (options.labels) {
      apiOptions.extra_labels = options.labels;
      delete apiOptions.labels;
    }
    if (options.imagePullSecret) {
      apiOptions.imagePullSecret = options.imagePullSecret;
    }
    if (options.workingDir) {
      apiOptions.workingDir = options.workingDir;
    }
    
    const timeoutParam = options.timeout;
    if (timeoutParam) {
      delete apiOptions.timeout; 
    }
    
    const params: Record<string, string> = {};
    if (timeoutParam) {
      params.timeout = timeoutParam;
    }

      const responseData = await super.post<BoxCreateResponse>(
        API_PREFIX,
        apiOptions,
        params,
        undefined, // headers
        signal
      );

      if (responseData.code! === 'ImagePullInProgress') {
        return responseData;
      }
      return this.mapLabels(responseData);
  }

  /**
   * Delete a specific box.
   * DELETE /api/v1/boxes/{id}
   */
  async deleteBox(
    boxId: string,
    force: boolean = false,
    signal?: AbortSignal
  ): Promise<{ message: string }> {
    const data = force ? { force } : undefined;
    return super.delete<{ message: string }>(
      `${API_PREFIX}/${boxId}`,
      data,
      undefined,
      undefined,
      signal
    );
  }

  /**
   * Delete all boxes.
   * DELETE /api/v1/boxes
   */
  async deleteAll(
    force: boolean = false,
    signal?: AbortSignal
  ): Promise<BoxesDeleteResponse> {
    const data = force ? { force } : undefined;
    return super.delete<BoxesDeleteResponse>(
      API_PREFIX,
      data,
      undefined,
      undefined,
      signal
    );
  }

  /**
   * Start a specific box.
   * POST /api/v1/boxes/{id}/start
   */
  async start(
    boxId: string,
    signal?: AbortSignal
  ): Promise<{ success: boolean; message: string }> {
    return super.post<{ success: boolean; message: string }>(
      `${API_PREFIX}/${boxId}/start`,
      {},
      undefined,
      undefined,
      signal
    );
  }

  /**
   * Stop a specific box.
   * POST /api/v1/boxes/{id}/stop
   */
  async stop(
    boxId: string,
    signal?: AbortSignal
  ): Promise<{ success: boolean; message: string }> {
    return super.post<{ success: boolean; message: string }>(
      `${API_PREFIX}/${boxId}/stop`,
      {},
      undefined,
      undefined,
      signal
    );
  }

  /**
   * Run a command in a box.
   * POST /api/v1/boxes/{id}/run
   */
  async run(
    boxId: string,
    command: string[],
    options?: BoxRunOptions,
    signal?: AbortSignal
  ): Promise<BoxRunResponse> {
    // Extract data payload fields from options
    const data = {
      cmd: command,
      ...(options?.stdin && { stdin: options.stdin }),
      ...(options?.stdoutLineLimit !== undefined && {
        stdout_line_limit: options.stdoutLineLimit,
      }),
      ...(options?.stderrLineLimit !== undefined && {
        stderr_line_limit: options.stderrLineLimit,
      }),
    };

    // Pass data payload and request config (with signal) separately
    const response = await super.post<BoxRunResponse>(
      `${API_PREFIX}/${boxId}/run`,
      data,
      undefined,
      undefined,
      signal
    );

    if (response.box) {
      response.box = this.mapLabels(response.box);
    }
    return response;
  }

  /**
   * Reclaim resources for a specific box or all inactive boxes.
   * POST /api/v1/boxes/reclaim
   * POST /api/v1/boxes/{id}/reclaim
   */
  async reclaim(
    boxId?: string,
    force: boolean = false,
    signal?: AbortSignal
  ): Promise<BoxReclaimResponse> {
    const data = { force };
    const url = boxId
      ? `${API_PREFIX}/${boxId}/reclaim`
      : `${API_PREFIX}/reclaim`;
    return super.post<BoxReclaimResponse>(
      url,
      data,
      undefined,
      undefined,
      signal
    );
  }

  /**
   * Get files from a box as a tar archive.
   * GET /api/v1/boxes/{id}/archive
   */
  async getArchive(
    boxId: string,
    path: string,
    signal?: AbortSignal
  ): Promise<ArrayBuffer> {
    const params = { path };
    return super.getRaw(
      `${API_PREFIX}/${boxId}/archive`,
      params,
      { Accept: 'application/x-tar' },
      signal
    );
  }

  /**
   * Extract a tar archive to a box.
   * PUT /api/v1/boxes/{id}/archive
   */
  async extractArchive(
    boxId: string,
    path: string,
    archiveData: ArrayBuffer,
    signal?: AbortSignal
  ): Promise<BoxExtractArchiveResponse> {
    const params = { path };
    return super.putRaw<BoxExtractArchiveResponse>(
      `${API_PREFIX}/${boxId}/archive`,
      archiveData,
      params,
      { 'Content-Type': 'application/x-tar' },
      signal
    );
  }

  /**
   * Get metadata about files in a box.
   * HEAD /api/v1/boxes/{id}/archive
   */
  async headArchive(
    boxId: string,
    path: string,
    signal?: AbortSignal
  ): Promise<Record<string, string>> {
    const params = { path };
    return super.head(
      `${API_PREFIX}/${boxId}/archive`,
      params,
      undefined,
      signal
    );
  }

  /**
   * Execute a command in a box via WebSocket, wait for completion, and return buffered output.
   * Connects via GET /ws/boxes/{id}/exec?cmd=...&arg=...&tty=...
   *
   * @param boxId The ID of the box.
   * @param cmd The command and its arguments as an array of strings.
   * @param options Optional settings like tty mode and abort signal.
   * @returns An object containing Readable streams for stdout and stderr, and a Promise for the exit code.
   */
  async exec(
    boxId: string,
    cmd: string[],
    options?: BoxExecOptions
  ): Promise<BoxExecProcess> {
    const { tty = false, signal, workingDir, stdin } = options ?? {};

    if (!cmd || cmd.length === 0) {
      throw new Error('cmd must be a non-empty array');
    }

    const wsUrlString = this.buildExecWsUrl(boxId, { cmd, tty, workingDir });

    let stdoutController!: ReadableStreamDefaultController;
    const stdoutStream = new ReadableStream({
      start(controller) {
        stdoutController = controller;
      },
    });

    let stderrController!: ReadableStreamDefaultController;
    const stderrStream = new ReadableStream({
      start(controller) {
        stderrController = controller;
      },
    });

    let resolveExitCode!: (code: number) => void;
    let rejectExitCode!: (reason?: any) => void;
    const exitCodePromise = new Promise<number>((resolve, reject) => {
      resolveExitCode = resolve;
      rejectExitCode = reject;
    });

    this._initiateWebSocketConnection({
      wsUrlString,
      tty,
      signal,
      stdin,
      stdoutController,
      stderrController,
      resolveExitCode,
      rejectExitCode,
    });

    return {
      stdout: stdoutStream,
      stderr: stderrStream,
      exitCode: exitCodePromise,
    };
  }

  private async _handleStdin(ws: WebSocketClient, input: string | ReadableStream): Promise<void> {
    logger.debug('Starting stdin handling.');
    try {
      if (typeof input === 'string') {
        const encoder = new TextEncoder();
        const data = encoder.encode(input);
        if (data.length > 0) {
          const copyData = data.slice();
          await ws.send(copyData.buffer as ArrayBuffer);
          logger.debug(`[WS] Sent ${data.length} bytes from string stdin.`);
        }
      } else if (input instanceof ReadableStream) {
        const reader = input.getReader();
        while (true) {
          const { done, value } = await reader.read();
          if (done) {
            logger.debug('[WS] stdin stream finished.');
            break;
          }
          if (value && value.length > 0) {
            if (value instanceof Uint8Array) {
              const copyData = value.slice();
              await ws.send(copyData.buffer as ArrayBuffer);
            } else if (value instanceof ArrayBuffer) {
              const copyData = new Uint8Array(value).slice();
              await ws.send(copyData.buffer as ArrayBuffer);
            } else {
              const data = new TextEncoder().encode(String(value));
              await ws.send(data.buffer as ArrayBuffer);
            }
            logger.debug(`[WS] Sent data from stdin stream.`);
          }
        }
        reader.releaseLock();
      }
      logger.debug('Finished writing stdin. Sending stdin_eof control message.');
      const eofMsg = JSON.stringify({ type: 'stdin_eof' });
      await ws.send(eofMsg);
    } catch (error: any) {
      logger.error('[WS] Error writing to stdin via WebSocket:', error);
    }
  }

  private _processWebSocketMessage(
    data: ArrayBuffer | string,
    tty: boolean,
    stdoutController: ReadableStreamDefaultController,
    stderrController: ReadableStreamDefaultController,
    frameBufferState: { buffer: Uint8Array },
    onExitCodeReceived: (code: number) => void
  ): void {
    if (typeof data === 'string') {
      try {
        const jsonMessage = JSON.parse(data);
        const exitMsg = jsonMessage as ExitMessage;
        if (exitMsg?.type === 'exit' && typeof exitMsg.exitCode === 'number') {
          logger.debug(`[WS] Received exit message: Code ${exitMsg.exitCode}`);
          onExitCodeReceived(exitMsg.exitCode);
        } else {
          logger.warn(`[WS] Received unexpected JSON text message:`, jsonMessage);
        }
      } catch (e) {
        logger.warn(`[WS] Received non-JSON text message: ${data}`);
      }
    } else if (data instanceof ArrayBuffer) {
      if (data.byteLength > 0) {
        if (tty) {
          stdoutController.enqueue(new Uint8Array(data));
        } else {
          const newData = new Uint8Array(data);
          const combined = new Uint8Array(frameBufferState.buffer.length + newData.length);
          combined.set(frameBufferState.buffer, 0);
          combined.set(newData, frameBufferState.buffer.length);
          frameBufferState.buffer = combined;

          frameBufferState.buffer = this._processDockerStreamFrames(
            frameBufferState.buffer,
            stdoutController,
            stderrController
          );
        }
      }
    } else {
      logger.warn(`[WS] Received message of unknown type: ${typeof data}`, data);
    }
  }

  private _processDockerStreamFrames(
    currentBuffer: Uint8Array,
    stdoutController: ReadableStreamDefaultController,
    stderrController: ReadableStreamDefaultController
  ): Uint8Array {
    let buffer = currentBuffer;
    logger.debug(`[WS] Processing ${buffer.length} bytes in _processDockerStreamFrames.`);
    while (buffer.length >= 8) {
      const header = buffer.slice(0, 8);
      const dataView = new DataView(header.buffer, header.byteOffset, header.byteLength);
      const streamType = dataView.getUint8(0) as StreamType;
      const payloadSize = dataView.getUint32(4, false); // Read as big-endian
      const frameSize = 8 + payloadSize;

      logger.debug(`[WS] Docker frame header: streamType=${streamType}, payloadSize=${payloadSize}, frameSize=${frameSize}, current buffer size=${buffer.length}`);

      if (buffer.length >= frameSize) {
        const payload = buffer.slice(8, frameSize);
        if (payload.byteLength > 0) {
          if (streamType === StreamTypeStdout) {
            logger.debug(`[WS] Enqueuing ${payload.byteLength} bytes to stdout stream.`);
            stdoutController.enqueue(payload);
          } else if (streamType === StreamTypeStderr) {
            logger.debug(`[WS] Enqueuing ${payload.byteLength} bytes to stderr stream.`);
            stderrController.enqueue(payload);
          } else {
            logger.warn(`[WS] Unknown stream type: ${streamType}`);
          }
        }
        buffer = buffer.slice(frameSize);
      } else {
        logger.debug(`[WS] Buffer (${buffer.length} bytes) too small for complete frame (needs ${frameSize} bytes). Waiting for more data.`);
        break;
      }
    }
    if (buffer.length > 0) {
      logger.debug(`[WS] Exiting _processDockerStreamFrames with ${buffer.length} bytes remaining in buffer.`);
    }
    return buffer;
  }

  private _initiateWebSocketConnection(params: {
    wsUrlString: string;
    tty: boolean;
    signal?: AbortSignal;
    stdin?: string | ReadableStream;
    stdoutController: ReadableStreamDefaultController;
    stderrController: ReadableStreamDefaultController;
    resolveExitCode: (code: number) => void;
    rejectExitCode: (reason?: any) => void;
  }): void {
    const {
      wsUrlString,
      tty,
      signal,
      stdin,
      stdoutController,
      stderrController,
      resolveExitCode,
      rejectExitCode,
    } = params;

    new Promise<void>((resolveWsLifecycle, rejectWsLifecycle) => {
      let receivedExitCode: number | null = null;
      let frameBufferState = { buffer: new Uint8Array(0) };
      let wsClientInstance: WebSocketClient | null = null;

      wsClientInstance = new WebSocketClient(
        wsUrlString,
        {
          signal: signal,
          onOpen: () => {
            logger.debug('[WS] WebSocket connection opened.');
            if (stdin && wsClientInstance) {
              this._handleStdin(wsClientInstance, stdin);
            }
          },
          onMessage: (data: ArrayBuffer | string) => {
            this._processWebSocketMessage(
              data,
              tty,
              stdoutController,
              stderrController,
              frameBufferState,
              (code) => {
                receivedExitCode = code;
              }
            );
          },
          onError: (error: Error) => {
            logger.error('[WS] WebSocket error:', error);
            stdoutController.error(error);
            stderrController.error(error);
            rejectExitCode(error);
            rejectWsLifecycle(error);
          },
          onClose: (code: number, reason: string, wasClean: boolean) => {
            logger.debug(
              `[WS] WebSocket closed. Code: ${code}, Reason: ${reason}, WasClean: ${wasClean}`
            );

            if (!tty && frameBufferState.buffer.length > 0) {
              logger.warn(
                `[WS] WebSocket closed with ${frameBufferState.buffer.length} unprocessed bytes in frame buffer.`
              );
            }

            if (receivedExitCode !== null) {
              resolveExitCode(receivedExitCode);
              stdoutController.close();
              stderrController.close();
              resolveWsLifecycle();
            } else {
              const closeError = new Error(
                `[WS] WebSocket closed (Code: ${code}, Reason: ${reason}, Clean: ${wasClean}) before receiving exit code.`
              );
              stdoutController.error(closeError);
              stderrController.error(closeError);
              rejectExitCode(closeError);
              rejectWsLifecycle(closeError);
            }
          },
        }
      );

      wsClientInstance.connect().catch((initialError) => {
        logger.error('[WS] Initial WebSocket connection failed:', initialError);
        stdoutController.error(initialError);
        stderrController.error(initialError);
        rejectExitCode(initialError);
        rejectWsLifecycle(initialError);
      });
    }).catch((wsError) => {
      logger.debug('[WS] WebSocket lifecycle promise rejected:', wsError.message);
    });
  }

  // Helper to construct the full WebSocket URL for the exec command
  private buildExecWsUrl(boxId: string, params: { cmd: string[]; tty: boolean; workingDir?: string }): string {
    const httpUrl = this.httpClient.defaults.baseURL;
    if (!httpUrl) {
      throw new Error('[WS] Cannot determine WebSocket URL: Axios baseURL is not set.');
    }

    const { cmd, tty, workingDir } = params;

    // Use API_PREFIX for the base path, as the route is under the API
    const wsBasePath = `${API_PREFIX}/${boxId}/exec/ws`;
    const url = new URL(wsBasePath, httpUrl); // Use httpUrl to resolve potential relative paths

    // Switch protocol from HTTP(S) to WS(S)
    url.protocol = url.protocol.replace(/^http/, 'ws');

    // Build search parameters
    // Ensure 'cmd' is treated as the command and subsequent elements as 'arg'
    if (cmd.length > 0) {
      url.searchParams.set('cmd', cmd[0]);
      cmd.slice(1).forEach(arg => url.searchParams.append('arg', arg));
    }
    if (tty) {
      url.searchParams.set('tty', 'true');
    }
    if (workingDir) {
      url.searchParams.set('workingDir', workingDir);
    }

    return url.toString();
  }

  // Helper to map extra_labels from API to labels in SDK consistently
  // Ensure input/output types are correct (T should extend BoxCreateResponse potentially)
  private mapLabels<T extends Partial<BoxData> & { extra_labels?: Record<string, string>; message?: string; }>(data: T): T & { labels?: Record<string, string> } {
    if (data && data.extra_labels) {
      data.labels = { ...(data.labels || {}), ...data.extra_labels };
      delete data.extra_labels;
    }
    return data;
  }
}
