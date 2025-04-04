import { Client } from "./client";
import type {
  Box,
  BoxRunResult,
  CreateBoxOptions,
  GBoxConfig,
  Logger,
  RunOptions,
} from "./types";

export class BoxService {
  private readonly client: Client;
  private readonly logger?: Logger;
  private readonly defaultCmd = ["sleep", "infinity"];
  private readonly defaultStdoutLimit = 100;
  private readonly defaultStderrLimit = 100;

  constructor(client: Client, config: GBoxConfig) {
    this.client = client;
    this.logger = config.logger;
  }

  // Helper function to build query parameters for filters
  private buildQueryParams(
    filters: Array<{ field: string; operator: "="; value: string }>
  ): string {
    if (filters.length === 0) {
      return "";
    }

    return (
      "?" +
      filters
        .map((filter) => `filter=${filter.field}=${filter.value}`)
        .join("&")
    );
  }

  // Get or create a box
  async getOrCreateBox(options: CreateBoxOptions): Promise<string> {
    const { boxId, image, sessionId, signal } = options;

    // If boxId is provided, try to use existing box
    if (boxId) {
      const response = await this.getBoxes({ sessionId, signal, boxId });
      const box = response.boxes.find((b) => b.id === boxId);
      if (box) {
        if (box.status === "stopped") {
          await this.startBox(boxId, signal);
        }
        return boxId;
      }
    }

    // Try to reuse an existing box with matching image
    const response = await this.getBoxes({ sessionId, signal });
    const boxes = response.boxes;

    // Try to find a running box with matching image
    const matchingBox = boxes.find(
      (box) => box.image === image && box.status === "running"
    );
    if (matchingBox) {
      return matchingBox.id;
    }

    // Try to find and start a stopped box
    const stoppedBox = boxes.find(
      (box) => box.image === image && box.status === "stopped"
    );
    if (stoppedBox) {
      await this.startBox(stoppedBox.id, signal);
      return stoppedBox.id;
    }

    // Create a new box if no matching box found
    const newBox = await this.createBox(image, this.defaultCmd, {
      sessionId,
      signal,
    });
    return newBox.id;
  }

  // Create a new box
  private async createBox(
    image: string,
    cmd: string | string[],
    { sessionId, signal }: RunOptions
  ): Promise<Box> {
    let command: string;
    let commandArgs: string[] | undefined;

    if (Array.isArray(cmd)) {
      [command, ...commandArgs] = cmd;
    } else {
      command = cmd;
    }

    try {
      const response = await this.client.post("/boxes", {
        body: JSON.stringify({
          image,
          cmd: command,
          args: commandArgs,
          labels: sessionId ? { sessionId } : undefined,
        }),
        signal,
      });

      if (!response.ok) {
        const errorText = await response.text();
        this.logger?.error(
          "Failed to create box. Status: %d, Error: %s",
          response.status,
          errorText
        );
        throw new Error(
          `Failed to create box: ${response.status} ${response.statusText}`
        );
      }

      const result = await response.json();
      this.logger?.debug("Box created successfully: %o", result);
      return result as Box;
    } catch (error) {
      this.logger?.error("Error creating box: %o", error);
      throw error;
    }
  }

  // Run command in a box and return output
  async runInBox(
    id: string,
    command: string | string[],
    stdin: string = "",
    stdoutLineLimit: number = this.defaultStdoutLimit,
    stderrLineLimit: number = this.defaultStderrLimit,
    { signal }: RunOptions
  ): Promise<BoxRunResult> {
    let cmd: string;
    let args: string[] | undefined;

    if (Array.isArray(command)) {
      [cmd, ...args] = command;
    } else {
      cmd = command;
    }

    try {
      const response = await this.client.post(`/boxes/${id}/run`, {
        body: JSON.stringify({
          cmd: [cmd],
          args,
          stdin,
          stdoutLineLimit,
          stderrLineLimit,
        }),
        signal,
      });

      if (!response.ok) {
        const errorText = await response.text();
        this.logger?.error(
          "Failed to run command in box. Status: %d, Error: %s",
          response.status,
          errorText
        );
        throw new Error(
          `Failed to run command in box: ${response.status} ${response.statusText}`
        );
      }

      const result = await response.json();
      this.logger?.debug("Command executed successfully: %o", result);
      return result as BoxRunResult;
    } catch (error) {
      this.logger?.error("Error running command in box: %o", error);
      throw error;
    }
  }

  // List all boxes
  async getBoxes({
    signal,
    sessionId,
    boxId,
  }: RunOptions & { boxId?: string }): Promise<{ boxes: Box[] }> {
    const filters = [];

    if (boxId) {
      filters.push({
        field: "id",
        operator: "=" as const,
        value: boxId,
      });
    }

    if (sessionId) {
      filters.push({
        field: "label",
        operator: "=" as const,
        value: `sessionId=${sessionId}`,
      });
    }

    const response = await this.client.get(
      `/boxes${this.buildQueryParams(filters)}`,
      {
        signal,
      }
    );

    return response.json();
  }

  // Start a box
  async startBox(id: string, signal?: AbortSignal): Promise<void> {
    await this.client.post(`/boxes/${id}/start`, { signal });
  }

  // Get a specific box by ID
  async getBox(id: string, { signal, sessionId }: RunOptions): Promise<Box> {
    const filters = [];

    if (sessionId) {
      filters.push({
        field: "label",
        operator: "=" as const,
        value: `sessionId=${sessionId}`,
      });
    }

    const response = await this.client.get(
      `/boxes/${id}${this.buildQueryParams(filters)}`,
      {
        signal,
      }
    );
    const result = await response.json();
    return result as Box;
  }
}
