# Contributing to NetTestLab

Thank you for your interest in contributing to NetTestLab! This document provides guidelines and information for contributors.

## 🤝 How to Contribute

### Reporting Issues

1. Check existing issues to avoid duplicates
2. Use the issue templates when available
3. Provide clear reproduction steps
4. Include system information (OpenWrt version, router model, etc.)

### Feature Requests

1. Describe the use case and motivation
2. Explain the expected behavior
3. Consider the impact on existing functionality
4. Discuss implementation approaches if you have ideas

### Code Contributions

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/your-feature-name`
3. Make your changes following our coding standards
4. Add tests for new functionality
5. Ensure all tests pass
6. Update documentation as needed
7. Submit a pull request

## 🛠️ Development Setup

### Prerequisites

- Go 1.21 or later
- Protocol Buffers compiler (`protoc`)
- Buf CLI for protobuf management
- OpenWrt build environment (for package testing)

### Local Development

1. **Clone and setup:**
   ```bash
   git clone https://github.com/yourusername/nettestlab.git
   cd nettestlab
   go mod download
   ```

2. **Generate protobuf files:**
   ```bash
   buf generate
   ```

3. **Run tests:**
   ```bash
   go test ./...
   ```

4. **Build the server:**
   ```bash
   go build -o bin/nettestlab cmd/server/main.go
   ```

### Testing

#### Unit Tests
```bash
go test ./...
```

#### Integration Tests
```bash
# Requires OpenWrt router for testing
go test -tags=integration ./tests/...
```

#### Manual Testing
```bash
# Test WiFi auto-discovery
go run cmd/wifi-test/main.go -server router-ip:8080

# Test with different clients
go run clients/go/examples/basic/main.go
```

## 📝 Coding Standards

### Go Code Style

- Follow Go conventions and use `gofmt`
- Use meaningful variable and function names
- Write clear comments for exported functions
- Keep functions small and focused
- Handle errors appropriately

### Protobuf Style

- Use snake_case for field names
- Include clear documentation for services and messages
- Version your APIs (e.g., `nettestlab.v1`)

### Git Conventions

#### Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
type(scope): description

[optional body]

[optional footer]
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes
- `refactor`: Code refactoring
- `test`: Test additions/modifications
- `chore`: Maintenance tasks

Examples:
```
feat(wifi): add auto-discovery for wireless interfaces
fix(server): handle graceful shutdown on SIGTERM
docs(api): update gRPC service documentation
```

#### Branch Naming

- `feature/description` - New features
- `fix/description` - Bug fixes
- `docs/description` - Documentation updates
- `refactor/description` - Code refactoring

## 📁 Project Structure

```
nettestlab/
├── api/                    # Generated gRPC clients
├── clients/               # Client libraries and examples
│   ├── go/               # Go client library
│   ├── javascript/       # JS/TS client library
│   ├── python/           # Python client library
│   └── dart/             # Dart/Flutter client library
├── cmd/                   # Main applications
│   ├── server/           # gRPC server
│   ├── client/           # CLI client
│   └── wifi-test/        # Testing utilities
├── docs/                  # Documentation
├── internal/              # Internal Go packages
│   ├── network/          # Network control logic
│   ├── profiles/         # Network profiles
│   ├── server/           # gRPC implementations
│   └── config/           # Configuration management
├── openwrt/              # OpenWrt packaging
│   ├── Makefile          # Package definition
│   └── files/            # Package files
├── proto/                # Protocol Buffer definitions
├── scripts/              # Build and utility scripts
├── tests/                # Integration tests
└── tools/                # Development tools
```

## 🧪 Testing Guidelines

### Test Categories

1. **Unit Tests**: Test individual functions and methods
2. **Integration Tests**: Test component interactions
3. **End-to-End Tests**: Test complete workflows
4. **Performance Tests**: Test under load conditions

### Test Naming

```go
func TestFunctionName_Scenario_ExpectedBehavior(t *testing.T) {
    // Test implementation
}
```

### Test Structure

Use the Arrange-Act-Assert pattern:

```go
func TestResolveInterfaceName_WiFiKeyword_ReturnsActualInterface(t *testing.T) {
    // Arrange
    controller := &NetworkController{}
    
    // Act
    result, err := controller.resolveInterfaceName("wifi")
    
    // Assert
    assert.NoError(t, err)
    assert.Contains(t, result, "wl")
}
```

## 📋 Pull Request Process

### Before Submitting

1. ✅ All tests pass
2. ✅ Code follows style guidelines
3. ✅ Documentation is updated
4. ✅ Commit messages follow conventions
5. ✅ No merge conflicts

### PR Description Template

```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
- [ ] Unit tests added/updated
- [ ] Integration tests pass
- [ ] Manual testing completed

## Checklist
- [ ] Code follows style guidelines
- [ ] Self-review completed
- [ ] Documentation updated
- [ ] Tests added for new functionality
```

### Review Process

1. Automated checks must pass
2. At least one maintainer review required
3. All discussions resolved
4. No merge conflicts

## 🚀 Release Process

### Versioning

We use [Semantic Versioning](https://semver.org/):

- `MAJOR.MINOR.PATCH`
- Major: Breaking changes
- Minor: New features (backward compatible)
- Patch: Bug fixes (backward compatible)

### Release Steps

1. Update version in relevant files
2. Update CHANGELOG.md
3. Create release tag
4. Build and publish packages
5. Update documentation

## 📞 Getting Help

- **GitHub Issues**: Bug reports and feature requests
- **GitHub Discussions**: Questions and general discussion
- **Documentation**: Check docs/ folder for detailed guides

## 🏆 Recognition

Contributors will be recognized in:
- README.md contributors section
- Release notes for significant contributions
- Project documentation

Thank you for contributing to NetTestLab! 🎉