import {
    GBoxClient,
    Box,
    NotFoundError,
    BrowserPage,
    BrowserContext,
    type BoxRunResponse,
    type BoxRunOptions,
    type BoxListFilters,
    type BoxCreateOptions,
    type VisionScreenshotResult,
    type VisionScreenshotParams,
    type FileShareApiResponse,
    BoxBrowserManager,
    FileManager,
    GBoxFile,
    type ImagePullStatus,
} from "../../../sdk/typescript/src/index";

const GBOX_URL = process.env.GBOX_URL || 'http://localhost:28080';

export const FILE_SIZE_LIMITS = {
    TEXT: 1024 * 1024, // 1MB for text files
    BINARY: 5 * 1024 * 1024, // 5MB for binary files (images, audio)
} as const;

const gbox = new GBoxClient({ baseURL: GBOX_URL, logger: { level: 'none', transports: [] } });

export {
    gbox,
    Box,
    BoxBrowserManager,
    FileManager,
    GBoxFile,
    NotFoundError,
    BrowserPage,
    BrowserContext,
    type BoxRunResponse,
    type BoxRunOptions,
    type BoxListFilters,
    type BoxCreateOptions,
    type VisionScreenshotResult,
    type VisionScreenshotParams,
    type FileShareApiResponse,
    type ImagePullStatus,
};

