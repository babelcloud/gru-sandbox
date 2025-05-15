import { 
    Box, 
    BoxBrowserManager, 
    NotFoundError, 
    type BoxRunResponse, 
    type BoxRunOptions, 
    type BoxListFilters, 
    type BoxCreateOptions, 
    type ImagePullStatus 
} from './gbox.instance.js';

import { gbox } from './gbox.instance.js';

export class BoxService {
    private readonly defaultImage = 'babelcloud/gbox-playwright';
    private readonly defaultCmd = ["sleep", "infinity"];
    private readonly defaultWaitTimeoutSeconds = 120;

    /**
     * Retrieves a list of Boxes, optionally filtered by sessionId or specific boxId.
     */
    async getBoxes(options: { sessionId?: string; boxId?: string; signal?: AbortSignal }): Promise<{ boxes: Box[]; count: number }> {
        try {
            let boxes: Box[];
            if (options.boxId) {
                try {
                    const box = await gbox.boxes.get(options.boxId, options.signal);
                    boxes = [box];
                } catch (error: unknown) {
                    if (error instanceof NotFoundError) {
                        boxes = [];
                    } else {
                        throw error;
                    }
                }
            } else {
                let filters: BoxListFilters | undefined;
                if (options.sessionId) {
                    filters = { label: [`sessionId=${options.sessionId}`] };
                }
                boxes = await gbox.boxes.list(filters, options.signal);
            }

            return { boxes, count: boxes.length };
        } catch (error) {
            throw error;
        }
    }

    /**
     * Gets a specific Box by its ID.
     * Corresponds to the old getBox.
     */
    async getBox(id: string, options?: { signal?: AbortSignal; sessionId?: string }): Promise<Box> {
        try {
            const box = await gbox.boxes.get(id, options?.signal);
            if (options?.sessionId && box.labels?.sessionId !== options.sessionId) {
                throw new Error(`Box ${id} found but does not belong to session ${options.sessionId}`);
            }
            return box;
        } catch (error: unknown) {
            if (error instanceof NotFoundError || (error instanceof Error && error.message.includes('does not belong to session'))) {
            } else {
            }
            throw error;
        }
    }

    /**
     * Starts a specific Box by its ID.
     * Corresponds to the old startBox.
     */
    async startBox(id: string, signal?: AbortSignal): Promise<void> {
        try {
            const box = await gbox.boxes.get(id, signal);
            await box.start(signal);
        } catch (error) {
            throw error;
        }
    }

    /**
     * Helper method to wait for a box to reach 'running' state.
     */
    private async _waitForBoxReady(boxId: string, timeoutSeconds?: number): Promise<void> {
        const effectiveTimeout = (timeoutSeconds ?? this.defaultWaitTimeoutSeconds) * 1000;
        const pollInterval = 2000;
        const startTime = Date.now();

        return new Promise((resolve, reject) => {
            const checkStatus = async () => {
                if (Date.now() - startTime > effectiveTimeout) {
                    return reject(new Error(`Timeout waiting for box ${boxId} to become ready`));
                }

                try {
                    const box = await this.getBox(boxId);
                    if (box.status === 'running') {
                        return resolve();
                    } else if (['stopped', 'error', 'deleted', 'exited'].includes(box.status)) {
                        return reject(new Error(`Box ${boxId} entered terminal state '${box.status}' while waiting`));
                    } else {
                        setTimeout(checkStatus, pollInterval);
                    }
                } catch (error: unknown) {
                    if (error instanceof NotFoundError) {
                         return reject(new Error(`Box ${boxId} not found while waiting for ready status.`));
                    }
                    return reject(error);
                }
            };
            setTimeout(checkStatus, 0);
        });
    }

    /**
     * Gets an existing Box or creates a new one based on the provided options.
     * Corresponds to the old getOrCreateBox.
     */
    async getOrCreateBox(options: {
        boxId?: string;
        image?: string;
        sessionId?: string;
        signal?: AbortSignal;
        waitTimeoutSeconds?: number;
    }): Promise<{ boxId: string | null; imagePullStatus?: { inProgress: boolean; imageName: string; message: string } }> {
        const { boxId, image, sessionId, signal, waitTimeoutSeconds } = options;

        if (boxId) {
            try {
                const box = await this.getBox(boxId, { sessionId, signal });
                if (box.status === 'stopped') {
                    await this.startBox(boxId, signal);
                    await this._waitForBoxReady(boxId, waitTimeoutSeconds);
                }
                return { boxId };
            } catch (error: unknown) {
                 if (error instanceof NotFoundError || (error instanceof Error && error.message.includes('does not belong to session'))) {
                 } else {
                    throw error;
                 }
            }
        }

        const effectiveImage = image || this.defaultImage;
        const listOptions = { sessionId, signal };
        const { boxes } = await this.getBoxes(listOptions);

        const runningBox = boxes.find(
            (box) => box.image.split(':')[0] === effectiveImage && box.status === 'running'
        );
        if (runningBox) {
            return { boxId: runningBox.id };
        }

        const stoppedBox = boxes.find(
            (box) => box.image.split(':')[0] === effectiveImage && box.status === 'stopped'
        );
        if (stoppedBox) {
            await this.startBox(stoppedBox.id, signal);
            await this._waitForBoxReady(stoppedBox.id, waitTimeoutSeconds);
            return { boxId: stoppedBox.id };
        }

        // Create a new box with timeout
        const createOptions: BoxCreateOptions = {
          image: effectiveImage,
          cmd: this.defaultCmd[0],
          args: this.defaultCmd.slice(1),
          labels: options.sessionId ? { sessionId: options.sessionId } : undefined,
          timeout: '1000ms', // Use reasonable timeout for image pull
        };

        // We need to use 'any' here because the SDK type system hasn't been updated
        // to reflect the new behavior of returning imagePullStatus
        try {
            const box = await gbox.boxes.create(createOptions, signal);
            if (box.id) {
                await this._waitForBoxReady(box.id, waitTimeoutSeconds);
                return { boxId: box.id };
            } else {
                return { boxId: null };
            }
        } catch (error) {
            if (error instanceof Error && error.message.includes('ImagePullInProgress')) {
                return { boxId: null, imagePullStatus: { inProgress: true, imageName: createOptions.image, message: error.message } };
            }
            throw error;
        }
    }

    /**
     * Initializes the browser manager for a specific Box.
     * This provides access to browser context and page operations via the SDK's manager.
     * @param boxId The ID of the Box.
     * @returns A Promise resolving to a BoxBrowserManager instance.
     */
    async initBrowser(boxId: string, signal?: AbortSignal): Promise<BoxBrowserManager> {
        try {
            const box = await gbox.boxes.get(boxId, signal);
            const browserManager = box.initBrowser();
            return browserManager;
        } catch (error) {
            throw error;
        }
    }

     /**
     * Runs a command inside a specific Box.
     * Corresponds to the old runInBox.
     * NOTE: Uses Box model's simple run(cmd: string[]), ignoring stdin/limit/signal options.
     */
    async runInBox(
        id: string,
        command: string | string[],
        stdin: string,
        stdoutLineLimit: number,
        stderrLineLimit: number,
        context: {
            signal?: AbortSignal;
            sessionId?: string;
        } = {}
    ): Promise<BoxRunResponse> {
        try {
            const box = await gbox.boxes.get(id, context.signal);

            let cmdArray: string[];
            if (Array.isArray(command)) {
                 cmdArray = command;
             } else {
                 cmdArray = command.split(' ');
             }

             const runOptions: BoxRunOptions = {
                 stdin: stdin,
                 signal: context?.signal,
                 stdoutLineLimit: stdoutLineLimit,
                 stderrLineLimit: stderrLineLimit,
             };

             const result = await box.run(cmdArray, runOptions, context.signal);

            return result;
        } catch (error) {
            throw error;
        }
    }
}
