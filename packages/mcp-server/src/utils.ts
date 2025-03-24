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
      try {
        log({ level: "info", data: `Starting ${handler.name}` });

        const result = await handler(log, ...args);

        log({ level: "info", data: `Completed ${handler.name} successfully` });

        return result;
      } catch (error) {
        log({
          level: "error",
          data: `Error in ${handler.name}: ${
            error instanceof Error ? error.message : String(error)
          }`,
        });
        throw error;
      }
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
        options.list(log,extra),
    };

    // Create and return the ResourceTemplate instance
    return new ResourceTemplate(uri, wrappedOptions);
  };
}
