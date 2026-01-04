# Testing Guide

This document describes how to test the frkr-operator.

## Testing Strategy

The operator uses a multi-layered testing approach:

1. **Unit Tests**: Test individual functions and methods with mocked dependencies
2. **Integration Tests**: Test controllers with a real Kubernetes API server (envtest)
3. **E2E Tests**: Test the full operator in a real cluster (optional, for CI/CD)

## Prerequisites

### Required Tools

```bash
# Install controller-gen (if not already installed)
go install sigs.k8s.io/controller-tools/cmd/controller-gen@latest

# Install kubebuilder (for envtest)
# On Linux:
curl -L -o kubebuilder https://go.kubebuilder.io/dl/latest/$(go env GOOS)/$(go env GOARCH)
chmod +x kubebuilder && sudo mv kubebuilder /usr/local/bin/

# Or use the Makefile target (see below)
```

### Test Dependencies

The test dependencies are already in `go.mod`. Key packages:
- `sigs.k8s.io/controller-runtime/pkg/envtest` - Kubernetes API server for testing
- `github.com/onsi/ginkgo/v2` - BDD testing framework (optional)
- `github.com/onsi/gomega` - Matcher library (optional)
- `github.com/stretchr/testify` - Alternative testing framework

## Running Tests

### Run All Tests

```bash
make test
# or
go test ./...
```

### Run Tests with Coverage

```bash
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Run Specific Test Package

```bash
go test ./internal/controller/... -v
```

### Run Specific Test

```bash
go test ./internal/controller -run TestDataPlaneReconciler -v
```

## Test Structure

### Unit Tests

Unit tests are co-located with the code they test:
- `internal/controller/dataplane_controller_test.go`
- `internal/controller/user_controller_test.go`
- etc.

### Integration Tests

Integration tests use `envtest` to spin up a real Kubernetes API server:

```go
import (
    "sigs.k8s.io/controller-runtime/pkg/envtest"
)
```

## Writing Tests

### Example: Controller Unit Test

See `internal/controller/dataplane_controller_test.go` for a complete example.

Key patterns:
1. Use `client.Client` interface for dependency injection
2. Use `fake.NewClientBuilder()` for unit tests
3. Use `envtest.Environment` for integration tests
4. Test both success and error paths
5. Verify status updates and resource creation

### Test Data

Create test fixtures in `testdata/` directory:
- Sample CRD manifests
- Expected outputs
- Mock responses

## CI/CD Integration

Tests run automatically in CI (see `.github/workflows/ci.yml`).

The CI pipeline:
1. Runs `make generate` to ensure code is generated
2. Runs `golangci-lint` for static analysis
3. Runs `go test ./...` for all tests

## Debugging Tests

### Verbose Output

```bash
go test ./... -v
```

### Run Single Test

```bash
go test ./internal/controller -run TestDataPlaneReconciler -v
```

### Debug with Delve

```bash
dlv test ./internal/controller -- -test.run TestDataPlaneReconciler
```

## Best Practices

1. **Test Coverage**: Aim for >80% coverage on critical paths
2. **Table-Driven Tests**: Use for testing multiple scenarios
3. **Test Isolation**: Each test should be independent
4. **Cleanup**: Always clean up test resources
5. **Error Cases**: Test both success and failure paths
6. **Status Updates**: Verify status is updated correctly
7. **RBAC**: Test that RBAC permissions are correct

## Common Test Scenarios

### Controller Tests Should Cover:

1. **Create**: Resource creation and status updates
2. **Update**: Resource updates and reconciliation
3. **Delete**: Resource deletion and cleanup
4. **Error Handling**: Invalid inputs, missing dependencies
5. **Status**: Status updates and conditions
6. **Reconciliation**: Multiple reconciliation cycles
7. **Edge Cases**: Empty values, nil pointers, etc.

## Troubleshooting

### envtest Issues

If envtest fails to start:
```bash
# Ensure kubebuilder is installed
which kubebuilder

# Check KUBEBUILDER_ASSETS environment variable
echo $KUBEBUILDER_ASSETS
```

### Test Timeouts

If tests timeout:
- Increase timeout: `go test -timeout 10m ./...`
- Check for resource leaks
- Verify cleanup in test teardown

### Flaky Tests

If tests are flaky:
- Add retries for eventual consistency
- Use `Eventually()` from gomega for async operations
- Check for race conditions

