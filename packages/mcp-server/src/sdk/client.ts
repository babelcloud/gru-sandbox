import type { Logger, GBoxConfig } from "./types";

export interface ClientRequestOptions extends RequestInit {
  params?: Record<string, string>;
  responseType?: "json" | "text" | "arrayBuffer";
}

export class Client {
  private readonly baseUrl: string;
  private readonly logger?: Logger;

  constructor(config: GBoxConfig) {
    this.baseUrl = config.apiUrl;
    this.logger = config.logger;
  }

  private log(level: keyof Logger, message: string, ...args: any[]): void {
    this.logger?.[level](message, ...args);
  }

  async request(
    path: string,
    options: ClientRequestOptions = {}
  ): Promise<Response> {
    const { params, ...restOptions } = options;
    const url = new URL(`${this.baseUrl}${path}`);

    if (params) {
      Object.entries(params).forEach(([key, value]) => {
        url.searchParams.append(key, value);
      });
    }

    // Log request
    this.log("debug", `${restOptions.method || "GET"} ${url.toString()}`);

    if (restOptions.body) {
      this.log("debug", `Request Body: ${restOptions.body}`);
    }

    const response = await fetch(url.toString(), {
      ...restOptions,
      headers: {
        "Content-Type": "application/json",
        ...restOptions.headers,
      },
      signal: restOptions.signal,
    });

    // Log response
    this.log(
      response.ok ? "debug" : "warn",
      `Response: ${response.status} ${response.statusText}`
    );

    return response;
  }

  async get(path: string, options: ClientRequestOptions = {}): Promise<Response> {
    return this.request(path, { ...options, method: "GET" });
  }

  async head(path: string, options: ClientRequestOptions = {}): Promise<Response> {
    return this.request(path, { ...options, method: "HEAD" });
  }

  async post(path: string, options: ClientRequestOptions = {}): Promise<Response> {
    return this.request(path, { ...options, method: "POST" });
  }

  async delete(path: string, options: ClientRequestOptions = {}): Promise<Response> {
    return this.request(path, { ...options, method: "DELETE" });
  }
}
