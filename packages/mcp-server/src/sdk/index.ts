export * from "./types";
export * from "./client";
export * from "./box";
export * from "./file";

import { BoxService } from "./box";
import { Client } from "./client";
import { FileService } from "./file";
import type { GBoxConfig } from "./types";

export class GBox {
  readonly client: Client;
  readonly box: BoxService;
  readonly file: FileService;

  constructor(config: GBoxConfig) {
    this.client = new Client(config);
    this.box = new BoxService(this.client, config);
    this.file = new FileService(this.client, config);
  }
}
