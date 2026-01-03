# frkr-operator

Kubernetes operator for managing the Traffic Mirroring Platform configuration.

## Purpose

This operator manages platform configuration via Kubernetes Custom Resource Definitions (CRDs):
- `FrkrUser` - User provisioning
- `FrkrAuthConfig` - Auth configuration (basic ↔ OIDC)
- `FrkrDataPlane` - Data plane configuration (BYO CockroachDB/Redpanda)
- `FrkrInit` - Database initialization (replaces frkr-init-core-stack)

## Requirements

- Go 1.21+ (required for Kubernetes client libraries)
- Kubernetes cluster for deployment
- kubectl configured

## Structure

```
frkr-operator/
├── cmd/
│   └── operator/
│       └── main.go
├── internal/
│   ├── controller/          # CRD controllers
│   │   ├── user.go
│   │   ├── auth.go
│   │   ├── datapane.go
│   │   ├── ingress.go
│   │   └── init.go
│   ├── reconciler/          # Reconciliation logic
│   └── k8s/                 # Kubernetes client utilities
├── api/
│   └── v1/
│       ├── user_types.go
│       ├── auth_types.go
│       ├── datapane_types.go
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
- Password reset support
- Auth configuration switching (deletes basic auth users on switch)
- Data plane configuration (validates connectivity, warns on errors)
- Ingress configuration (Envoy required, auto-configured, BYO certs)
- Database initialization (runs migrations via golang-migrate)

## License

Apache 2.0
