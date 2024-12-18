# RunShell

A secure command execution framework designed for AI/LLM agents to safely interact with system operations.

[![English](https://img.shields.io/badge/README-English-blue)](README.md) [![中文](https://img.shields.io/badge/README-中文-red)](README_zh.md)

![CI Status](https://github.com/iamlongalong/runshell/workflows/CI/badge.svg)
![Release Status](https://github.com/iamlongalong/runshell/workflows/Release/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/iamlongalong/runshell)](https://goreportcard.com/report/github.com/iamlongalong/runshell)
[![codecov](https://codecov.io/gh/iamlongalong/runshell/branch/main/graph/badge.svg)](https://codecov.io/gh/iamlongalong/runshell)

## Features

- **AI/LLM Integration**
  - AI-friendly command interface
  - Secure execution environment
  - Perfect for LLM tool chains
  - Comprehensive audit logging

- **Execution Modes**
  - Local command execution
  - Docker container isolation
  - Interactive shell
  - HTTP API service

- **Security Features**
  - Command execution auditing
  - User permission control
  - Resource usage monitoring
  - Timeout control

- **Additional Features**
  - Environment variable management
  - Working directory control
  - I/O stream handling
  - Error management

## Quick Start

### Installation

#### From Source

```bash
# Clone repository
git clone https://github.com/iamlongalong/runshell.git
cd runshell

# Install dependencies
make deps

# Build project
make build-local
```

#### Using Docker

```bash
# Build and run Docker container
make docker-build docker-run
```

### Usage

#### Command Line

```bash
# Execute simple command
runshell exec -- ls -l

# Set working directory
runshell exec --workdir /tmp -- ls -l

# Set environment variables
runshell exec --env KEY=VALUE env

# Example of using Docker image
runshell exec --docker-image ubuntu:latest -- ls -l
runshell exec --docker-image busybox:latest --env KEY=VALUE env
runshell exec --docker-image busybox:latest --workdir /app -- python3 script.py

# Start HTTP server
runshell server --http :8080
```

#### HTTP API Examples

```bash
# Health check
curl http://localhost:8080/api/v1/health

# Execute command
curl -X POST http://localhost:8080/api/v1/exec \
  -H "Content-Type: application/json" \
  -d '{
    "command": "ls",
    "args": ["-l"],
    "workdir": "/tmp",
    "env": {"KEY": "VALUE"}
  }'

# List available commands
curl http://localhost:8080/api/v1/commands

# Get command help
curl http://localhost:8080/api/v1/help?command=ls

# Session Management
# Create new session, info: now does not support docker_config, only support options
curl -X POST http://localhost:8080/api/v1/sessions \
  -H "Content-Type: application/json" \
  -d '{
    "executor_type": "docker",
    "docker_config": {
      "image": "golang:1.20",
      "workdir": "/workspace",
      "bind_mount": "/local/path:/workspace"
    },
    "options": {
      "workdir": "/workspace",
      "env": {"GOPROXY": "https://goproxy.cn,direct"}
    }
  }'

# List all sessions
curl http://localhost:8080/api/v1/sessions

# Execute command in session
curl -X POST http://localhost:8080/api/v1/sessions/{session_id}/exec \
  -H "Content-Type: application/json" \
  -d '{
    "command": "ls",
    "args": ["-al"],
    "options": {
      "workdir": "/"
    }
  }'

# Delete session
curl -X DELETE http://localhost:8080/api/v1/sessions/{session_id}

# Interactive shell (WebSocket)
# npm install -g wscat
wscat -c ws://localhost:8080/api/v1/exec/interactive
# after connected, you can use the following commands:
# ls -al
# exit
```

## Development Guide

### Make Commands

The project provides a series of Make commands to simplify development and deployment:

#### Basic Operations
- `make` - Run default operations (clean, test, build)
- `make clean` - Clean build artifacts
- `make deps` - Update dependencies
- `make help` - Show all available commands

#### Testing
- `make test` - Run all tests
- `make test-unit` - Run unit tests only
- `make coverage` - Generate coverage report

#### Building
- `make build` - Build for all platforms
- `make build-local` - Build for current platform only

#### Docker Operations
- `make docker-build` - Build Docker image
- `make docker-run` - Run Docker container
- `make docker-stop` - Stop Docker container

#### Development Tools
- `make fmt` - Format code
- `make lint` - Check code style
- `make run` - Run local server
- `make tag` - Create new Git tag

### Release Process

1. Create Release Candidate:
   ```bash
   # Create RC tag
   git tag -a v1.0.0-rc.1 -m "Release candidate 1 for version 1.0.0"
   git push origin v1.0.0-rc.1
   ```

2. Test Release Candidate:
   - GitHub Actions will automatically:
     - Create pre-release GitHub Release
     - Build and upload binaries
     - Build and push Docker RC image (`:rc` tag)
   - Download and test pre-release version
   - If issues found, fix and repeat steps 1-2, incrementing RC version

3. Create Official Release:
   ```bash
   # Create release tag
   git tag -a v1.0.0 -m "Release version 1.0.0"
   git push origin v1.0.0
   ```

### Docker Images

Docker images are available on Docker Hub:

```bash
# Use latest stable version
docker pull iamlongalong/runshell:latest

# Use pre-release version
docker pull iamlongalong/runshell:rc

# Use specific version
docker pull iamlongalong/runshell:v1.0.0
```

## Contributing

1. Fork the project
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Create a Pull Request

Before submitting code, please ensure:

1. All tests pass (`make test`)
2. Code meets standards (`make lint`)
3. Documentation is updated
4. Test cases are added

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Author

- iamlongalong

## Acknowledgments

Thanks to the following open source projects:

- [cobra](https://github.com/spf13/cobra)
- [docker](https://github.com/docker/docker) 