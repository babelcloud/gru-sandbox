export * from "./types";
export * from "./client";
export * from "./box";
export * from "./file";

import { Client } from "./client";
import { BoxService } from "./box";
import { FileService } from "./file";
import type { SDKConfig } from "./types";

export class GBox {
  readonly client: Client;
  readonly box: BoxService;
  readonly file: FileService;

  constructor(config: SDKConfig) {
    this.client = new Client(config);
    this.box = new BoxService(this.client);
    this.file = new FileService(this.client);
  }
}
