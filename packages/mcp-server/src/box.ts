// API server configuration
const API_SERVER_URL = "http://localhost:28080/api/v1";

// HTTP client for API calls
async function apiRequest(path: string, options: RequestInit = {}) {
  const response = await fetch(`${API_SERVER_URL}${path}`, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      ...options.headers,
    },
    signal: options.signal,
  });

  if (!response.ok) {
    throw new Error(`API request failed: ${response.statusText}`);
  }

  return response.json();
}

// Box interface
interface Box {
  id: string;
  image: string;
  status: string;
  labels?: Record<string, string>;
  // Add other box properties as needed
}

// Box list response interface
interface BoxListResponse {
  boxes: Box[];
}

type BoxOptions = {
  signal?: AbortSignal;
  sessionId?: string;
};

type RunOptions = {
  boxId?: string;
} & BoxOptions;

// Get all boxes
export async function getBoxes({
  signal,
  sessionId,
}: BoxOptions): Promise<Box[]> {
  const queryParams = sessionId ? `?filter=label=sessionId=${sessionId}` : "";
  const response = (await apiRequest(`/boxes${queryParams}`, {
    signal,
  })) as BoxListResponse;
  return response.boxes;
}

// Get a single box by ID
export async function getBox(
  id: string,
  { signal, sessionId }: BoxOptions
): Promise<Box> {
  const queryParams = sessionId ? `?filter=label=sessionId=${sessionId}` : "";
  return apiRequest(`/boxes/${id}${queryParams}`, { signal });
}

// Create a new box
export async function createBox(
  image: string = "ubuntu:latest",
  command: string[] = ["/bin/bash"],
  { sessionId, signal }: BoxOptions
): Promise<Box> {
  const extraLabels = sessionId ? { sessionId } : undefined;
  return apiRequest("/boxes", {
    method: "POST",
    body: JSON.stringify({ image, command, extraLabels }),
    signal,
  });
}

// Start a stopped box
export async function startBox(id: string, signal?: AbortSignal): Promise<Box> {
  return apiRequest(`/boxes/${id}/start`, {
    method: "POST",
    signal,
  });
}

type GetOrCreateBoxOptions = {
  boxId?: string;
  image?: string;
} & BoxOptions;

// Get or create a box with specific image
export async function getOrCreateBox({
  boxId,
  image,
  sessionId,
  signal,
}: GetOrCreateBoxOptions): Promise<string> {
  if (boxId) {
    const boxes = await getBoxes({ sessionId, signal });
    const box = boxes.find((b) => b.id === boxId);
    if (box) {
      // If box exists but is stopped, start it
      if (box.status === "stopped") {
        await startBox(boxId, signal);
      }
      return boxId;
    }
  }

  // If no boxId provided, try to reuse an existing box with matching image
  const boxes = await getBoxes({ sessionId, signal });

  // First try to find a running box with matching image
  const runningBox = boxes.find(
    (box) => box.image === image && box.status === "running"
  );
  if (runningBox) {
    return runningBox.id;
  }

  // Then try to find a stopped box with matching image
  const stoppedBox = boxes.find(
    (box) => box.image === image && box.status === "stopped"
  );
  if (stoppedBox) {
    await startBox(stoppedBox.id, signal);
    return stoppedBox.id;
  }

  // If no matching box found, create a new one
  const response = await createBox(image, ["/bin/bash"], {
    sessionId,
    signal,
  });
  return response.id;
}

// Run command in a box and return output
export async function runInBox(
  id: string,
  command: string[],
  args: string[] = [],
  stdin: string = "",
  stdoutLineLimit: number = 100,
  stderrLineLimit: number = 100,
  { sessionId, signal }: RunOptions
): Promise<{
  exit_code: number;
  stdout: string;
  stderr: string;
}> {
  const queryParams = sessionId ? `?filter=label=sessionId=${sessionId}` : "";
  return apiRequest(`/boxes/${id}/run${queryParams}`, {
    method: "POST",
    body: JSON.stringify({
      cmd: command,
      args,
      stdin,
      stdoutLineLimit,
      stderrLineLimit,
    }),
    signal,
  });
}

// Run Python code in a box
export async function runPython(
  code: string,
  { boxId, sessionId, signal }: RunOptions
): Promise<{
  exit_code: number;
  stdout: string;
  stderr: string;
}> {
  const id = await getOrCreateBox({
    boxId,
    image: "python:3.13-bookworm",
    sessionId,
    signal,
  });
  // Read code from stdin and execute it
  return runInBox(id, ["python3"], [], code, 100, 100, { sessionId, signal });
}

// Run Bash command in a box
export async function runBash(
  code: string,
  { boxId, sessionId, signal }: RunOptions
): Promise<{
  exit_code: number;
  stdout: string;
  stderr: string;
}> {
  const id = await getOrCreateBox({
    boxId,
    image: "ubuntu:latest",
    sessionId,
    signal,
  });
  // Read code from stdin and execute it
  return runInBox(id, ["/bin/bash"], [], code, 100, 100, { sessionId, signal });
}
