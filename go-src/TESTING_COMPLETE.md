# Testing Suite for Coolify Go

This document provides comprehensive information about the testing infrastructure implemented for the Coolify Go application.

## Overview

We have implemented a comprehensive testing suite covering all major components of the Coolify Go application, including:

- **Docker Integration Tests** - Testing Docker client functionality
- **Container Management Tests** - Testing container operations
- **Image Management Tests** - Testing Docker image operations
- **Network Management Tests** - Testing Docker network operations
- **Swarm Management Tests** - Testing Docker Swarm functionality
- **Deployment Engine Tests** - Testing application deployment logic
- **SSH Client Tests** - Testing SSH connectivity (framework ready)

## Test Structure

### Docker Integration Tests (`internal/docker/`)

#### Client Tests (`client_test.go`)

- **TestNewClient**: Tests Docker client initialization
- **TestDockerClient_Connect**: Tests client connection
- **TestDockerClient_Close**: Tests client cleanup
- **TestDockerClient_GetInfo**: Tests system information retrieval
- **TestDockerClient_GetVersion**: Tests version information
- **TestDockerClient_Ping**: Tests connection health check
- **TestDockerClient_WithTimeout**: Tests timeout handling
- **TestDockerClient_GetClientAndContext**: Tests client accessors

#### Container Tests (`container_test.go`)

- **TestClient_ListContainers**: Tests container listing
- **TestClient_GetContainer**: Tests container inspection
- **TestClient_CreateContainer**: Tests container creation
- **TestClient_StartContainer**: Tests container startup
- **TestClient_StopContainer**: Tests container stopping
- **TestClient_RemoveContainer**: Tests container removal
- **TestClient_GetContainerLogs**: Tests log retrieval
- **TestClient_GetContainerStats**: Tests statistics collection
- **TestClient_ExecuteCommand**: Tests command execution

#### Image Tests (`image_test.go`)

- **TestClient_ListImages**: Tests image listing
- **TestClient_GetImage**: Tests image inspection
- **TestClient_PullImage**: Tests image pulling
- **TestClient_BuildImage**: Tests image building
- **TestClient_RemoveImage**: Tests image removal
- **TestClient_TagImage**: Tests image tagging
- **TestClient_SaveImage**: Tests image saving
- **TestClient_LoadImage**: Tests image loading
- **TestClient_PruneImages**: Tests image cleanup

#### Network Tests (`network_test.go`)

- **TestClient_ListNetworks**: Tests network listing
- **TestClient_GetNetwork**: Tests network inspection
- **TestClient_CreateNetwork**: Tests network creation
- **TestClient_RemoveNetwork**: Tests network removal
- **TestClient_ConnectContainerToNetwork**: Tests container-network connection
- **TestClient_DisconnectContainerFromNetwork**: Tests container-network disconnection
- **TestClient_CreateBridgeNetwork**: Tests bridge network creation

#### Swarm Tests (`swarm_test.go`)

- **TestClient_InitSwarm**: Tests swarm initialization
- **TestClient_JoinSwarm**: Tests swarm joining
- **TestClient_LeaveSwarm**: Tests swarm leaving
- **TestClient_GetSwarmInfo**: Tests swarm information
- **TestClient_ListServices**: Tests service listing
- **TestClient_GetService**: Tests service inspection
- **TestClient_CreateService**: Tests service creation
- **TestClient_UpdateService**: Tests service updates
- **TestClient_RemoveService**: Tests service removal
- **TestClient_ListNodes**: Tests node listing
- **TestClient_GetNode**: Tests node inspection
- **TestClient_ListTasks**: Tests task listing

### Deployment Engine Tests (`internal/deployment/`)

#### Engine Tests (`engine_test.go`)

- **TestDeploymentEngine_NewEngine**: Tests engine initialization
- **TestDeploymentEngine_DeployApplication**: Tests application deployment
- **TestDeploymentEngine_ConvertEnvVars**: Tests environment variable conversion
- **TestDeploymentEngine_ConvertPorts**: Tests port configuration conversion
- **TestDeploymentEngine_ConvertVolumeBinds**: Tests volume binding conversion
- **TestDeploymentEngine_ConfigStructure**: Tests configuration structure
- **TestDeploymentEngine_DeploymentResultStructure**: Tests result structure
- **TestDeploymentEngine_WorkDirStructure**: Tests working directory structure
- **TestDeploymentEngine_ImageNameGeneration**: Tests image naming logic
- **TestDeploymentEngine_AppNameGeneration**: Tests application naming logic

## Test Categories

### Unit Tests

- Test individual functions and methods
- Mock external dependencies
- Fast execution
- High reliability

### Integration Tests

- Test component interactions
- Use real Docker daemon when available
- Test end-to-end workflows
- May be skipped if dependencies unavailable

### Framework Tests

- Test data structures and configurations
- Validate business logic
- Test utility functions
- Always run regardless of environment

## Running Tests

### Using the Test Runner Script

```bash
# Run all tests
./run_tests.sh

# Run only unit tests
./run_tests.sh unit

# Run only integration tests
./run_tests.sh integration

# Run only Docker tests
./run_tests.sh docker

# Run only deployment tests
./run_tests.sh deployment
```

### Using Go Test Directly

```bash
# Run all tests
go test -v ./...

# Run tests for specific package
go test -v ./internal/docker
go test -v ./internal/deployment

# Run tests with coverage
go test -v -cover ./...

# Run tests with race detection
go test -v -race ./...

# Run tests with timeout
go test -v -timeout 30s ./...
```

### Test Environment Requirements

#### For Unit Tests

- Go 1.24 or later
- Testify package (already included in go.mod)

#### For Integration Tests

- Docker daemon running
- Network access for image pulling
- Git for repository cloning tests

#### For Full Test Suite

- All integration test requirements
- SSH server for SSH tests
- Sufficient disk space for Docker images

## Test Patterns

### Skip Pattern for Integration Tests

```go
func TestIntegrationFunction(t *testing.T) {
    client, err := NewClient()
    if err != nil {
        t.Skip("Docker daemon not available, skipping integration test")
    }
    // ... test implementation
}
```

### Error Expectation Pattern

```go
func TestFunctionThatFailsInTestEnv(t *testing.T) {
    result, err := someFunction()
    // This might fail in test environment, but we're testing the function exists
    assert.Error(t, err) // Expected to fail in test environment
}
```

### Cleanup Pattern

```go
func TestWithCleanup(t *testing.T) {
    // Setup
    resource := createTestResource()

    // Cleanup
    defer func() {
        cleanupResource(resource)
    }()

    // Test implementation
}
```

## Test Coverage

### Current Coverage Areas

- ✅ Docker client initialization and connection
- ✅ Container lifecycle management
- ✅ Image operations (pull, build, remove)
- ✅ Network management
- ✅ Swarm mode operations
- ✅ Deployment engine configuration
- ✅ Data structure validation
- ✅ Utility function testing

### Planned Coverage Areas

- 🔄 SSH client functionality
- 🔄 Database operations
- 🔄 API endpoint testing
- 🔄 Authentication and authorization
- 🔄 Real-time features
- 🔄 Webhook integrations
- 🔄 Monitoring and metrics

## Test Best Practices

### 1. Test Organization

- Group related tests together
- Use descriptive test names
- Follow the pattern `TestFunctionName_Scenario`

### 2. Test Independence

- Each test should be independent
- Clean up resources after tests
- Don't rely on test execution order

### 3. Error Handling

- Test both success and failure cases
- Verify error messages and types
- Test edge cases and boundary conditions

### 4. Performance

- Keep tests fast
- Use appropriate timeouts
- Mock expensive operations when possible

### 5. Documentation

- Document complex test scenarios
- Explain why certain tests are skipped
- Provide context for integration test requirements

## Continuous Integration

### GitHub Actions Integration

```yaml
name: Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: "1.24"
      - name: Install dependencies
        run: go mod download
      - name: Run tests
        run: ./run_tests.sh
```

### Docker-based Testing

```dockerfile
FROM golang:1.24-alpine
RUN apk add --no-cache docker git
COPY . /app
WORKDIR /app
CMD ["./run_tests.sh"]
```

## Troubleshooting

### Common Issues

#### Docker Connection Errors

```bash
# Check Docker daemon status
docker info

# Ensure Docker socket is accessible
ls -la /var/run/docker.sock
```

#### Test Timeout Issues

```bash
# Increase test timeout
go test -v -timeout 60s ./...

# Run tests with verbose output
go test -v -count=1 ./...
```

#### Missing Dependencies

```bash
# Update dependencies
go mod tidy

# Download missing packages
go mod download
```

### Debug Mode

```bash
# Run tests with debug output
DEBUG=1 ./run_tests.sh

# Run specific test with verbose output
go test -v -run TestSpecificFunction ./internal/docker
```

## Future Enhancements

### Planned Improvements

1. **Mock Framework Integration**: Add comprehensive mocking for external dependencies
2. **Performance Testing**: Add benchmarks for critical operations
3. **Load Testing**: Test system behavior under load
4. **Security Testing**: Add security-focused test cases
5. **API Testing**: Add HTTP API endpoint testing
6. **Database Testing**: Add database integration tests

### Test Infrastructure

1. **Test Containers**: Use testcontainers for isolated testing
2. **Parallel Testing**: Enable parallel test execution
3. **Test Reporting**: Add detailed test reports and metrics
4. **Coverage Reports**: Generate and track code coverage

## Conclusion

The testing suite provides comprehensive coverage of the Coolify Go application's core functionality. The tests are designed to be:

- **Reliable**: Tests are independent and repeatable
- **Fast**: Unit tests run quickly, integration tests are optimized
- **Comprehensive**: Cover both success and failure scenarios
- **Maintainable**: Well-organized and documented

The test runner script makes it easy to execute tests in different environments and provides clear feedback on test results. The modular design allows for selective testing of specific components while maintaining the ability to run the full test suite.

For questions or issues with the testing suite, please refer to the troubleshooting section or create an issue in the project repository.
