import axios from 'axios'; // Default import for runtime value
// Import types separately using 'import type'
import type {
  AxiosInstance,
  AxiosError,
  AxiosResponseHeaders,
  RawAxiosResponseHeaders,
  AxiosResponse,
} from 'axios';
import {
  APIError,
  NotFoundError,
  GBoxError,
  ConflictError,
} from '../errors.ts';
import { logger } from '../logger.ts'; // Import the logger

export class Client {
  protected httpClient: AxiosInstance;

  // Accept logger in constructor, provide default
  constructor(httpClient: AxiosInstance) { // logger parameter removed
    this.httpClient = httpClient;
  }

  // Centralized error handler (can be used as an interceptor or called directly)
  private handleError(error: any): never {
    // Use 'never' as it always throws
    logger.error('API Request Error:', error); // Commented out to avoid logging the full raw error object
    if (axios.isAxiosError(error)) {
      const axiosError = error as AxiosError;
      const statusCode = axiosError.response?.status;
      const responseData = axiosError.response?.data;
      // Try to get message from response data first, then axios message
      const message =
        (responseData as any)?.message ||
        axiosError.message ||
        'An Axios error occurred';

      const specificErrorMap: {
        [key: number]: new (msg: string, data?: any) => APIError;
      } = {
        404: NotFoundError,
        409: ConflictError,
        // Add other specific status codes here if needed
      };

      const ErrorClass = statusCode ? specificErrorMap[statusCode] : undefined;

      if (ErrorClass) {
        logger.warn(
          `Request failed (${ErrorClass.name}): ${message}`,
          responseData
        );
        throw new ErrorClass(message, responseData);
      } else if (statusCode && statusCode >= 400 && statusCode < 600) {
        logger.error(
          `Request failed (APIError ${statusCode}): ${message}`,
          responseData
        );
        throw new APIError(message, statusCode, responseData);
      } else {
        logger.error(
          `Request failed (Network/Unknown Axios Error): ${message}`,
          responseData
        );
        throw new APIError(message, undefined, responseData);
      }
    } else {
      logger.error(
        `Unexpected non-Axios error: ${(error as Error).message}`,
        error
      );
      throw new GBoxError(
        `An unexpected error occurred: ${(error as Error).message}`
      );
    }
  }

  // Core request method - can be protected if only helpers are public
  protected async request<T>(config: {
    method: string;
    url: string;
    data?: any;
    params?: any;
    headers?: Record<string, string>;
    responseType?: 'json' | 'arraybuffer';
    signal?: AbortSignal;
  }): Promise<AxiosResponse<T>> {
    logger.debug(
      `[http] Requesting: ${config.method.toUpperCase()} ${config.url}`,
      { params: config.params, data: config.data }
    ); // Log request details
    try {
      const response = await this.httpClient.request<T>({
        ...config,
        signal: config.signal,
      });
      logger.debug(`[http] Response: ${response.status} ${response.statusText}`, {
        url: config.url,
      }); // Log success response status
      return response;
    } catch (error) {
      this.handleError(error); // Delegate error handling
    }
  }

  // --- Public Helper Methods --- (Mimicking Python Client)

  async get<T = any>(
    path: string,
    params?: Record<string, any>,
    headers?: Record<string, string>,
    signal?: AbortSignal
  ): Promise<T> {
    const response = await this.request<T>({
      method: 'get',
      url: path,
      params,
      headers,
      responseType: 'json',
      signal,
    });
    return response.data;
  }

  async getRaw(
    path: string,
    params?: Record<string, any>,
    headers?: Record<string, string>,
    signal?: AbortSignal
  ): Promise<ArrayBuffer> {
    // Explicitly set Accept header if needed, e.g., application/octet-stream or x-tar
    const response = await this.request<ArrayBuffer>({
      method: 'get',
      url: path,
      params,
      headers,
      responseType: 'arraybuffer',
      signal,
    });
    return response.data;
  }

  async post<T = any>(
    path: string,
    data?: any,
    params?: Record<string, any>,
    headers?: Record<string, string>,
    signal?: AbortSignal
  ): Promise<T> {
    const response = await this.request<T>({
      method: 'post',
      url: path,
      data,
      params,
      headers,
      responseType: 'json',
      signal,
    });
    return response.data;
  }

  async put<T = any>(
    path: string,
    data?: any,
    params?: Record<string, any>,
    headers?: Record<string, string>,
    signal?: AbortSignal
  ): Promise<T> {
    const response = await this.request<T>({
      method: 'put',
      url: path,
      data,
      params,
      headers,
      responseType: 'json',
      signal,
    });
    return response.data;
  }

  async putRaw<T = any>(
    path: string,
    data: ArrayBuffer,
    params?: Record<string, any>,
    headers?: Record<string, string>,
    signal?: AbortSignal
  ): Promise<T> {
    // Content-Type header is crucial here and should be passed in headers
    // Assume the response IS expected to be JSON by default for putRaw, caller specifies T
    const response = await this.request<T>({
      method: 'put',
      url: path,
      data,
      params,
      headers,
      responseType: 'json',
      signal,
    });
    return response.data;
  }

  async delete<T = any>(
    path: string,
    data?: any,
    params?: Record<string, any>,
    headers?: Record<string, string>,
    signal?: AbortSignal
  ): Promise<T> {
    const response = await this.request<T>({
      method: 'delete',
      url: path,
      data,
      params,
      headers,
      responseType: 'json',
      signal,
    });
    return response.data;
  }

  async head(
    path: string,
    params?: Record<string, any>,
    headers?: Record<string, string>,
    signal?: AbortSignal
  ): Promise<Record<string, string>> {
    logger.debug(`[http] Requesting: HEAD ${path}`, { params });
    try {
      // Use httpClient directly for HEAD as it doesn't typically have a body to parse
      const response = await this.httpClient.head(path, {
        params,
        headers,
        signal,
      });
      logger.debug(`[http] Response: ${response.status} ${response.statusText}`, {
        url: path,
      });
      // Convert Axios headers to simple Record<string, string>
      return this.convertHeaders(response.headers);
    } catch (error) {
      // Error logging handled by handleError
      this.handleError(error);
    }
  }

  private convertHeaders(
    headers: RawAxiosResponseHeaders | AxiosResponseHeaders | undefined
  ): Record<string, string> {
    const result: Record<string, string> = {};
    if (!headers) return result;

    // Axios headers can be complex; iterate safely
    for (const key in headers) {
      // Check for own properties, Axios headers might have methods etc.
      if (Object.prototype.hasOwnProperty.call(headers, key)) {
        const value = headers[key];
        if (typeof value === 'string') {
          result[key.toLowerCase()] = value;
        } else if (Array.isArray(value)) {
          result[key.toLowerCase()] = value.join(', ');
        }
        // Ignore number/boolean headers if any exist
      }
    }
    return result;
  }
}
