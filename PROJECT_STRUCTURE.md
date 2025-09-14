# Project Structure

```
NetTestLab/
├── ARCHITECTURE.md          # System architecture documentation
├── DEVELOPMENT_PLAN.md      # Development roadmap and milestones  
├── README.md                # Project overview and quick start
├── LICENSE                  # MIT License
├── Makefile                 # Build automation
├── go.mod                   # Go module definition
├── buf.yaml                 # Buf configuration
├── buf.gen.yaml             # Buf code generation config
├── mkdocs.yml               # Documentation configuration
├── .gitignore               # Git ignore patterns
│
├── cmd/                     # Main applications
│   └── server/              # gRPC server entry point
│
├── internal/                # Private application code
│   ├── network/             # Network control logic
│   ├── server/              # gRPC server implementation
│   └── profiles/            # Network profile management
│
├── pkg/                     # Public library code
│   └── client/              # Go client library
│
├── proto/                   # Protocol Buffer definitions
│
├── api/                     # Generated gRPC code (auto-generated)
│
├── clients/                 # Client libraries for different languages
│   ├── javascript/          # JavaScript/TypeScript client
│   ├── python/              # Python client  
│   ├── java/                # Java client
│   ├── dart/                # Dart/Flutter client
│   └── go/                  # Go client
│
├── openwrt/                 # OpenWRT package
│   └── package/             # Package definition and scripts
│
├── docs/                    # Documentation (MkDocs)
│   ├── index.md             # Documentation home page
│   ├── api/                 # API reference docs
│   ├── installation/        # Installation guides
│   └── examples/            # Usage examples
│
├── scripts/                 # Build and utility scripts
│   └── build-openwrt-package.sh  # OpenWrt package build script
│
└── .github/                 # GitHub Actions workflows
    └── workflows/           # CI/CD pipeline definitions
```

## Key Directories

### Core Application (`cmd/`, `internal/`, `pkg/`)
- `cmd/server/` - Main gRPC server application
- `internal/network/` - Network control using tc/netem  
- `internal/server/` - gRPC service implementations
- `internal/profiles/` - Network condition presets
- `pkg/client/` - Reusable Go client library

### Protocol Definitions (`proto/`, `api/`)
- `proto/` - Source Protocol Buffer (.proto) files
- `api/` - Generated Go code from protobuf (auto-generated)

### Client Libraries (`clients/`)
Each subdirectory contains a complete client package for the respective language:
- Language-specific packaging (package.json, setup.py, pom.xml, etc.)
- Generated code from Protocol Buffers
- Language-specific documentation and examples

### OpenWRT Integration (`openwrt/`)
- `package/` - OpenWRT package definition
- Makefile for cross-compilation
- Installation and configuration scripts

### Documentation (`docs/`)
- MkDocs-based documentation
- API reference generated from protobuf
- Installation guides for all platforms
- Usage examples and tutorials

### Automation (`.github/`, `scripts/`)
- GitHub Actions for CI/CD
- Build scripts for local development
- Publishing workflows for package registries