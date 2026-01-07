# frkr-operator

Kubernetes operator for managing frkr configuration.

## Purpose

This operator manages platform configuration via Kubernetes Custom Resource Definitions (CRDs):
- `FrkrUser` - User provisioning
- `FrkrStream` - Stream management (retention, description)
- `FrkrAuthConfig` - Auth configuration (basic ↔ OIDC)
- `FrkrDataPlane` - Data plane configuration (BYO PostgreSQL-compatible DB and Kafka-compatible broker)
- `FrkrInit` - Database initialization (replaces frkr-init-core-stack)

## Requirements

- Go 1.21+ (required for Kubernetes client libraries)
- Kubernetes cluster for deployment
- kubectl configured

## Structure

```
frkr-operator/
├── cmd/
│   ├── operator/            # Operator entry point
│   │   └── main.go
│   └── frkrctl/             # CLI for managing CRDs
│       └── main.go
├── internal/
│   ├── controller/          # CRD controllers
│   │   ├── user_controller.go
│   │   ├── auth_controller.go
│   │   ├── dataplane_controller.go
│   │   ├── stream_controller.go
│   │   └── init_controller.go
│   ├── reconciler/          # Reconciliation logic
│   └── k8s/                 # Kubernetes client utilities
├── api/
│   └── v1/
│       ├── user_types.go
│       ├── auth_types.go
│       ├── datapane_types.go
│       ├── stream_types.go
│       └── init_types.go
├── config/
│   └── crd/                 # CRD manifests
├── Dockerfile
├── go.mod
└── README.md
```

## Dependencies

- `frkr-common` - Shared library (plugin interfaces, migrations)
- Kubernetes operator framework (controller-runtime)

## Features

- User provisioning with password generation (one-time display in status)
- Stream management (create/update/delete Kafka topics and DB entries)
- Password reset support
- Auth configuration switching (deletes basic auth users on switch)
- Data plane configuration (validates connectivity, warns on errors)
- Ingress configuration (Envoy required, auto-configured, BYO certs)
- Database initialization (runs migrations via golang-migrate)

## Building

See [BUILD.md](BUILD.md) for detailed build instructions, including:
- Prerequisites (Go 1.21+, controller-gen)
- Code generation steps (DeepCopy methods)
- Troubleshooting common issues

## Testing

See [TESTING.md](TESTING.md) for comprehensive testing guide, including:
- Unit test examples
- Integration test setup
- Running tests
- Test coverage

**Quick Start**:
```bash
# Install controller-gen
go install sigs.k8s.io/controller-tools/cmd/controller-gen@latest
export PATH=$PATH:$(go env GOPATH)/bin

# Generate code
make generate

# Build
make build
```

## License

Apache 2.0
