# gru-sandbox

Gru-sandbox(gbox) is an open source project that provides a self-hostable sandbox for MCP integration or other AI agent usecases.

As MCP is getting more and more popular, we find there is no easy way to enable MCP client such as Claude Desktop/Cursor to execute commands locally and securely. This project is based on the technology behind [gru.ai](https://gru.ai) and we wrap it into a system command and MCP server to make it easy to use.

For advanced scenarios, we also kept the ability to run sandboxes in k8s cluster locally or remotely.

## Use Cases

Your AI client such as Claude Desktop can use gbox MCP to deliver better results, such as

### 1. Generating Diagrams

Generate diagrams of Tesla stock prices:
![Image](https://i.imghippo.com/files/njBB6977VQQ.png)
https://claude.ai/share/34de8ca3-4e04-441b-9e79-5875fa9fc97a

### 2. Generating PDFs

Generate PDF of latest AI news:
![Image](https://i.imghippo.com/files/oMF9723LA.png)
https://claude.ai/share/84600933-dcf2-44be-a2fd-7f49540db57a

### 3. Analyzing and Calculation

Analyze and compare Nvidia/Tesla market cap:
![Image](https://i.imghippo.com/files/FE2710WR.png)
https://claude.ai/share/70c335b7-9fff-4ee7-8459-e6b7462d8994

### 4. Processing Local Files (coming soon)

```
Please compress all photos in shared folder and make sure each of them is smaller than 2MB.
```

### 5. Execute Arbitrary Tasks

Download youtube video:
![Image](https://i.imghippo.com/files/TI9396Rjg.png)
https://claude.ai/share/c2ab6bcb-7032-489f-87d5-cc38f72c2ca9

## Installation

### System Requirements

- macOS 10.15 or later
- [Docker Desktop for Mac](https://docs.docker.com/desktop/setup/install/mac-install/)
- [Homebrew](https://brew.sh)

> Note: Support for other platforms (Linux, Windows) is coming soon.

### Installation Steps

```bash
# Install via Homebrew
brew tap babelcloud/gru && brew install gbox

# Initialize environment
gbox setup

# Export MCP config and merge into Claude Desktop
gbox mcp export --merge-to claude
# or gbox mcp export --merge-to cursor

# Restart Claude Desktop
```

### Update Steps

```bash
# Update gbox to the latest version
brew update && brew upgrade gbox

# Update the environment
gbox setup

# Export and merge latest MCP config into Claude Desktop
gbox mcp export --merge-to claude
# or gbox mcp export --merge-to cursor

# Restart Claude Desktop
```

## Command Line Usage

The project provides a command-line tool `gbox` for managing sandbox containers:

```bash
# Cluster management
gbox cluster setup    # Setup cluster environment
gbox cluster cleanup  # Cleanup cluster environment

# Container management
gbox box create --image python:3.9 --env "DEBUG=true" -w /app -v /host/path:/app   # Create container
gbox box list                                                                      # List containers
gbox box start <box-id>                                                            # Start container
gbox box stop <box-id>                                                             # Stop container
gbox box delete <box-id>                                                           # Delete container
gbox box exec <box-id> -- python -c "print('Hello')"                               # Execute command
gbox box inspect <box-id>                                                          # Inspect container

# MCP configuration
gbox mcp export                          # Export MCP configuration
gbox mcp export --merge-to claude        # Export and merge into Claude Desktop config
gbox mcp export --dry-run                # Preview merge result without applying changes
```

### Volume Mounts

The `gbox box create` command supports Docker-compatible volume mounts using the `-v` or `--volume` flag. This allows you to share files and directories between your host system and the sandbox containers.

The volume mount syntax follows this format:

```bash
-v /host/path:/container/path[:ro][:propagation]
```

Where:

- `/host/path`: Path to a file or directory on your host system
- `/container/path`: Path where the file or directory will be mounted in the container
- `ro` (optional): Makes the mount read-only
- `propagation` (optional): Sets the mount propagation mode (private, rprivate, shared, rshared, slave, rslave)

Examples:

```bash
# Basic bind mount
gbox box create -v /data:/data --image python:3.9

# Read-only bind mount
gbox box create -v /data:/data:ro --image python:3.9

# Multiple bind mounts
gbox box create \
  -v /config:/etc/myapp \
  -v /data:/var/lib/myapp:ro \
  -v /logs:/var/log/myapp:ro:rprivate \
  --image python:3.9
```

Note: The host path must exist before creating the container. The container path will be created automatically if it doesn't exist.

## Development Setup

### Prerequisites

- Go 1.21 or later
- Docker Desktop
- Make
- pnpm (via corepack)
- Node.js 16.13 or later

### Build

```bash
# Build all components
make build

# Create distribution package
make dist
```

### Running Services

```bash
# API Server
make -C packages/api-server dev

# MCP Server
cd packages/mcp-server && pnpm dev

# MCP Inspector
cd packages/mcp-server && pnpm inspect
```

## Contributing

We welcome contributions! Please feel free to submit a Pull Request. For major changes, please open an issue first to discuss what you would like to change.

1. Fork the repository
2. Create your feature branch (`git checkout -b username/feature-name`)
3. Commit your changes (`git commit -m 'Add some feature'`)
4. Push to the branch (`git push origin username/feature-name`)
5. Open a Pull Request

### Things to Know about Dev and Debug Locally

#### How to run gbox in dev env instead of the system installed one

1. Stop the installed gbox by `gbox cleanup`. It will stop the api server so that you can run the api server in dev env.
2. Execute `make api-dev` in project root.
3. Execute `./gbox box list`, this is the command run from your dev env.

#### How to connect MCP client such as Claude Desktop to the MCP server in dev env

1. Execute `make mcp-dev` in project root.
2. Execute `./gbox mcp export --merge-to claude`

#### How to open MCP inspect

1. Execute `make mcp-inspect` in project root.
2. Click the link returned in terminal.

#### How to build and use image in dev env

1. Execute `make build-image-python` in project root to build Python image, or `make build-images` to build all images.
2. Change the image name as needed (e.g., `make build-image-typescript` for TypeScript image).
3. You may need to delete current sandboxes to make the new image effective `./gbox box delete --all`

#### Why MCP client still get the old MCP content?

1. After you change MCP configuration such as tool definitions, you need to run `make build` to update the `dist/index.js` file.
2. You may also need to execute `./gbox mcp export --merge-to claude`

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.
