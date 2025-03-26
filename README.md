# gru-sandbox

Gru-sandbox(gbox) is an opensource project that provides a self-hostable sandbox for MCP integration or other AI agent usecases.

As MCP is getting more and more popular, we find there is no easy way to enable MCP client such as Claude Desktop/Cursor to execute commands locally and securely. This project is based on the technology behined [gru.ai](https://gru.ai) and we wrap it into a system command and MCP server to make it easy to use. 

For advanced scenarios, we also kept the ability to run sandboxes in k8s cluster locally or remotely. 

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
gbox box create --image python:3.9 --env "DEBUG=true" -w /app  # Create container
gbox box list                                                  # List containers
gbox box start <box-id>                                        # Start container
gbox box stop <box-id>                                         # Stop container
gbox box delete <box-id>                                       # Delete container
gbox box exec <box-id> -- python -c "print('Hello')"           # Execute command
gbox box inspect <box-id>                                      # Inspect container

# MCP configuration
gbox mcp export                # Export MCP configuration
gbox mcp export --merge        # Export and merge into Claude Desktop config
gbox mcp export --dry-run      # Preview merge result without applying changes
```

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
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.
