# k3sdemo

Multi-app k3s GitOps lab: Go, Node.js, Spring Boot, Laravel. Fleet for delivery, MetalLB for LoadBalancer IPs, Traefik as ingress.

## Stack

| App | Image | Namespace | Ingress |
|-----|-------|-----------|---------|
| godemo | `ghcr.io/alexuresp/godemo` | `godemo` | `http://godemo.192-168-1-151.traefik.me` |
| godemo-dev | `ghcr.io/alexuresp/godemo-dev` | `godemo-dev` | `http://godemo-dev.192-168-1-151.traefik.me` |
| node-api | `ghcr.io/alexuresp/node-api` | `node-api` | `http://node-api.192-168-1-151.traefik.me` |
| spring-app | `ghcr.io/alexuresp/spring-app` | `spring-app` | `http://spring-app.192-168-1-151.traefik.me` |
| laravel | `ghcr.io/alexuresp/laravel-app` | `laravel` | `http://laravel.192-168-1-151.traefik.me` |

## Layout

```
├── godemo/                  # Go web app
├── node-api/                # Node.js + Express
├── spring-app/              # Spring Boot 3 + Java 21
├── laravel-app/             # Laravel 11 + PHP-FPM + Nginx + Redis + PostgreSQL
├── fleet/
│   ├── godemo/
│   ├── godemo-dev/
│   ├── node-api/
│   ├── spring-app/
│   └── laravel-app/
├── examples/                # GitRepo manifests for Fleet
├── .github/workflows/       # CI: build → push → yq update → git push
├── metallb-pool.yaml        # MetalLB IPAddressPool + L2Advertisement
└── README.md
```

## Prerequisites

- k3s + Rancher (Fleet / Continuous Delivery)
- Public GitHub repo: https://github.com/Alexuresp/godemo
- `kubectl` and kubeconfig (`config` is gitignored)

## MetalLB

k3s ships with Klipper LB (`svclb-traefik`), which does not preserve the real client IP for LoadBalancer services. MetalLB replaces it with a proper L2 load balancer.

### Install

```bash
kubectl apply -f https://raw.githubusercontent.com/metallb/metallb/v0.14.5/config/manifests/metallb-native.yaml
```

### Configure IP pool

```bash
kubectl apply -f metallb-pool.yaml
```

Pool: `192.168.1.151-192.168.1.199` (outside DHCP range `192.168.1.100-150`).

### Verify

```bash
kubectl -n metallb-system get pods
kubectl -n kube-system get svc traefik
```

Expected output:
```
traefik   LoadBalancer   10.43.18.81   192.168.1.151   ...
```

### Real IP caveat

Even with MetalLB, in single-node k3s the real client IP is still not visible inside pods if traffic passes through Klipper LB host-port rules. This is a known limitation of the k3s ServiceLB architecture. For production, use an external load balancer or `hostNetwork: true` with caution (may break CoreDNS).

### Files

- `fleet/traefik/helmchartconfig.yaml` — `forwardedHeaders.insecure: true`, `externalTrafficPolicy: Local`
- `metallb-pool.yaml` — `IPAddressPool` + `L2Advertisement`

## Deploy flow

1. **Push to `main`** — GitHub Actions builds and pushes app images to GHCR.
2. **Register the GitRepo once per app:**

```bash
kubectl --kubeconfig config apply -f examples/gitrepo.yaml
kubectl --kubeconfig config apply -f examples/gitrepo-node-api.yaml
kubectl --kubeconfig config apply -f examples/gitrepo-spring-app.yaml
kubectl --kubeconfig config apply -f examples/gitrepo-laravel-app.yaml
```

3. **Check:**

```bash
kubectl --kubeconfig config -n fleet-local get gitrepo
kubectl --kubeconfig config -n godemo get deploy,svc,ingress,pods
kubectl --kubeconfig config -n node-api get deploy,svc,ingress,pods
kubectl --kubeconfig config -n spring-app get deploy,svc,ingress,pods
kubectl --kubeconfig config -n laravel get deploy,svc,ingress,pods
```

## CI per app

| Workflow | Branch | Image |
|----------|--------|-------|
| `.github/workflows/build.yml` | `main` → `godemo`, `develop` → `godemo-dev` | `ghcr.io/alexuresp/godemo`, `ghcr.io/alexuresp/godemo-dev` |
| `.github/workflows/node-api.yml` | `main` | `ghcr.io/alexuresp/node-api` |
| `.github/workflows/spring-app.yml` | `main` | `ghcr.io/alexuresp/spring-app` |
| `.github/workflows/laravel-app.yml` | `main` | `ghcr.io/alexuresp/laravel-app` |

## Production-grade details

- **Non-root containers**: all apps run as non-root users
- **Health checks**: readiness/liveness probes on `/healthz` or `/actuator/health`
- **Resources**: requests/limits set for every container
- **Secrets**: passwords and keys in Kubernetes Secrets, not in Git
- **ConfigMaps**: environment-specific config separated from code
- **Image tags**: SHA tags via CI, `latest` only on default branch
- **Multi-stage Docker**: minimal runtime images
- **Logs**: stdout/stderr for all containers, collectable via `kubectl logs`

## Local run

```bash
cd godemo && go run .
# http://localhost:8080
```

```bash
cd node-api && npm install && npm start
# http://localhost:3000
```

```bash
cd spring-app && mvn spring-boot:run
# http://localhost:8080
```

```bash
cd laravel-app && composer install && php artisan serve
# http://localhost:8000
```
