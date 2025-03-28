import { Server } from "@modelcontextprotocol/sdk/server/index.js";
import { ResourceTemplate } from "@modelcontextprotocol/sdk/server/mcp.js";
import { RequestHandlerExtra } from "@modelcontextprotocol/sdk/shared/protocol.js";
import type { ListResourcesResult } from "@modelcontextprotocol/sdk/types.js";

// Log function type from SDK
export type LogFunction = typeof Server.prototype.sendLoggingMessage;

export type WithLoggingHandler<T extends (...args: any[]) => any> = (
  log: LogFunction,
  ...args: Parameters<T>
) => ReturnType<T>;

// Wrapper function to add logging capability to any handler
export function withLogging<T extends (...args: any[]) => any>(
  handler: WithLoggingHandler<T>
): (log: LogFunction) => T {
  return (log: LogFunction) => {
    return (async (...args: Parameters<T>) => {
      return await handler(log, ...args);
    }) as T;
  };
}

// Resource template types
type ResourceTemplateCallback = ConstructorParameters<
  typeof ResourceTemplate
>[1];
type ListCallback = NonNullable<ResourceTemplateCallback["list"]>;

// Wrapper function specifically for ResourceTemplate
export function withLoggingResourceTemplate(
  uri: string,
  options: Omit<ResourceTemplateCallback, "list"> & {
    list: WithLoggingHandler<ListCallback>;
  }
): (log: LogFunction) => ResourceTemplate {
  return (log: LogFunction) => {
    // Create a new options object with wrapped list method
    const wrappedOptions = {
      ...options,
      list: async (extra: RequestHandlerExtra): Promise<ListResourcesResult> =>
        options.list(log, extra),
    };

    // Create and return the ResourceTemplate instance
    return new ResourceTemplate(uri, wrappedOptions);
  };
}

// This file is intentionally empty as the snake_case to camelCase conversion
// is no longer needed since the API now uses camelCase by default.
