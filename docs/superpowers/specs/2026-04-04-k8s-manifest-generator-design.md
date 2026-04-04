# K8s Manifest Generator Design

**Date:** 2026-04-04
**Status:** Approved

## Overview

A Go CLI tool that reads a YAML config file and generates a single multi-document Kubernetes YAML manifest. Opinionated for simplicity: sensible defaults, no templating, fully type-safe.

## CLI Interface

```
deployer generate <config-file> [-o <output-file>]
```

- `-o` defaults to `-` (stdout)
- Reads the config file, calls the generator, writes YAML to the output

A const `defaultIngressClass = "traefik"` lives in `main.go` and is passed into the generator via `Options`, making it injectable for future use.

## Package Structure

```
deployer/
├── main.go
├── config/
│   └── config.go            # existing — unchanged
└── generator/
    ├── generator.go         # Generate(cfg, opts) orchestrator
    ├── labels.go            # buildLabels(name, component)
    ├── namespace.go         # generateNamespace(cfg)
    ├── podspec.go           # buildPodSpec(r) + buildContainer(r) — shared
    ├── deployment.go        # generateDeployment(svc, cfg, opts)
    ├── service.go           # generateService(svc, cfg)
    ├── ingress.go           # generateIngress(svc, cfg, opts)
    └── job.go               # generateJob(job, cfg)
```

## Options

```go
type Options struct {
    IngressClass string
}
```

Constructed in `main.go` with `IngressClass: "traefik"`.

## Resource Mapping

Resources are emitted in this order: Namespace, then per-Service resources, then per-InitJob resources.

### Always

| Input | K8s Resource |
|---|---|
| `Config` | `Namespace` (name = `cfg.Name`) |

### Per `Service`

| Condition | K8s Resource |
|---|---|
| always | `Deployment` (1 replica) |
| `Ingress != nil` | `Service` (ClusterIP, port = `Ingress.Port`) |
| `Ingress.Public != nil` | `Ingress` (host = `Public.Domain`, class = `opts.IngressClass`, path = `/`, pathType = `Prefix`) |

### Per `InitJob`

| Condition | K8s Resource |
|---|---|
| always | `Job` (`backoffLimit` = `MaxRetries`, default 0 if nil) |

## Pod Spec (Shared)

`buildPodSpec(r config.Runnable) corev1.PodSpec` and `buildContainer(r config.Runnable) corev1.Container` are shared between `generateDeployment` and `generateJob`, since both embed `config.Runnable`.

Container spec from `Runnable`:
- `image`: `<Repository>:<Tag>`
- `command` + `args` mapped directly
- `env`: `rawValue` → `env[].value`; env entries where `RawValue` is nil are skipped (reserved for future types like `persistentRandom`)
- `resources.requests.cpu`: set, no limit
- `resources.requests.memory` + `resources.limits.memory`: both set

## Labels

Applied to all resources (metadata labels) and pod template labels:

| Label | Value |
|---|---|
| `app.kubernetes.io/name` | service or job name |
| `app.kubernetes.io/part-of` | `cfg.Name` |
| `app.kubernetes.io/managed-by` | `deployer` |

Deployment/Service selectors use `app.kubernetes.io/name` only.

`buildLabels(name, partOf string) map[string]string` in `labels.go` is the single source of truth, making it easy to extend in future.

## YAML Serialization

```go
func Generate(cfg config.Config, opts Options) ([]byte, error)
```

- Builds `[]runtime.Object` in resource order
- Each object marshaled with `sigs.k8s.io/yaml`
- Documents joined with `---\n`
- Each resource has `TypeMeta` set explicitly (apiVersion + kind)

## Dependencies

- `k8s.io/api` — `appsv1`, `batchv1`, `corev1`, `networkingv1`
- `k8s.io/apimachinery` — `ObjectMeta`, `TypeMeta`, resource quantities
- `sigs.k8s.io/yaml` — YAML marshaling

## Testing

Each generator function is unit-tested by asserting on the returned Go struct (no YAML parsing in unit tests).

Key unit test cases:
- `generateDeployment`: image, CPU request-only, memory request+limit, labels, env
- `generateJob`: backoffLimit from MaxRetries, nil MaxRetries defaults to 0
- `generateService`: only generated when Ingress != nil
- `generateIngress`: only generated when Public != nil, uses ingressClass
- `buildLabels`: correct label keys/values
- `buildPodSpec`: shared fields applied correctly to both Deployment and Job

One integration test in `generator_test.go` calls `Generate` with a full config and compares output against a golden file at `generator/testdata/expected.yaml`.
