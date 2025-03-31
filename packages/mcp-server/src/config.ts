import * as dotenv from "dotenv-defaults";
import * as dotenvExpand from "dotenv-expand";
import path from "path";
import { fileURLToPath } from "url";

// Get the directory name in ES modules
const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

// Load environment variables with defaults and expansion
const env = dotenv.config({
  path: path.resolve(__dirname, "../.env"),
  defaults: path.resolve(__dirname, "../.env.defaults"),
  multiline: true,
});
dotenvExpand.expand(env);

// Export environment variables with type safety
export const config = {
  apiServer: {
    url: process.env.API_SERVER_URL || "http://localhost:28080/api/v1",
  },
  images: {
    python: process.env.PY_IMG || "babelcloud/gbox-python:latest",
    typescript: process.env.TS_IMG || "babelcloud/gbox-typescript:latest",
    bash: process.env.SH_IMG || "babelcloud/gbox-python:latest",
  },
} as const;
