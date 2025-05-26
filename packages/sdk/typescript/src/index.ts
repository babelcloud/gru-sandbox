// Export the main client
export { GBoxClient } from './client.ts';

// Export configuration types/interfaces if needed for external use
export type { GBoxClientConfig /*, Logger */ } from './config.ts';

// Export managers if direct access is desired (though usually accessed via client)
// export { BoxManager } from './managers/boxManager.ts';
// export { FileManager } from './managers/fileManager.ts';
// export { BrowserManager } from './managers/browserManager.ts';

// Export custom error classes
export * from './errors.ts';

// Export types/interfaces that users of the SDK might need
export * from './types/box.ts';
export * from './types/file.ts';

// Export Model classes
export { Box } from './models/box.ts';
export { GBoxFile } from './models/file.ts';
export { BrowserContext } from './models/context.ts';
export { BrowserPage } from './models/page.ts';
export { BoxBrowserManager } from './managers/browser.manager.ts';
export { FileManager } from './managers/file.manager.ts';
export { GboxClient as SandboxClient } from './sandbox/client.ts';

// Export new browser-related exports
export * from './api/browser.api.ts';
export * from './types/browser.ts';

// Re-export SdkLogger and LogLevel for external use
export { logger } from './logger.ts';

// Export library components
export * from './client.ts';

// Export specific error classes (kept for backward compatibility)
export { ImagePullInProgressError } from './api/box.api.ts';
