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
├── godemo-dev/               # Go web app (dev)
├── node-api/                # Node.js + Express
├── spring-app/              # Spring Boot 3 + Java 21
├── laravel-app/             # Laravel 11 + PHP-FPM + Nginx + Redis + PostgreSQL
└── fleet/
    ├── godemo/
    ├── godemo-dev/
    ├── node-api/
    ├── spring-app/
    └── laravel-app/
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

## PostgreSQL backups

PostgreSQL is managed by **CloudNativePG** (v1.30.0). Each app namespace has its own CNPG `Cluster` resource, and backups are stored in **MinIO** via barman-cloud.

### Cluster layout

| Namespace | CNPG Cluster | Service RW | App credentials |
|-----------|--------------|------------|-----------------|
| `godemo` | `godemo` | `godemo-rw.godemo.svc.cluster.local:5432` | user `godemo`, DB `godemo` |
| `godemo-dev` | `godemo-dev` | `godemo-dev-rw.godemo-dev.svc.cluster.local:5432` | user `godemo-dev`, DB `godemo-dev` |
| `laravel` | `laravel` | `laravel-rw.laravel.svc.cluster.local:5432` | user `laravel`, DB `laravel` |

### Backup destination

- **MinIO endpoint:** `http://10.132.7.1:9000`
- **Bucket:** `cnpg-backups`
- **Paths:** `s3://cnpg-backups/godemo`, `s3://cnpg-backups/godemo-dev`, `s3://cnpg-backups/laravel`
- **Retention:** 7 days
- **Format:** compressed data backups + WAL archiving (continuous archiving enabled)

Backups run automatically from the CNPG operator. No separate CronJob is needed.

### Restore with `cnpg` plugin (recommended)

Install the plugin:

```bash
kubectl cnpg install
```

List available backups:

```bash
kubectl cnpg backup ls -n godemo
kubectl cnpg backup ls -n godemo-dev
kubectl cnpg backup ls -n laravel
```

Restore the latest backup to a new cluster (point-in-time recovery is also supported):

```bash
kubectl cnpg restore backup -n godemo --target-name godemo-restored
```

### Restore manually (from MinIO)

If you need to restore into an existing cluster:

```bash
# 1. List objects in MinIO bucket
mc ls myminio/cnpg-backups/godemo/

# 2. Download a specific backup
mc cp myminio/cnpg-backups/godemo/base_.../data.tar.gz ./data.tar.gz

# 3. Restore with pg_restore
kubectl exec -n godemo deploy/godemo-1 -- \
  pg_restore -U postgres -d godemo --clean /backups/data.tar.gz
```

### Important notes

- CNPG `Cluster` replaces the old `postgres` Deployment/Service/PVC. Old resources were removed during migration.
- App secrets (`DATABASE_URL`) now point to `*-rw` CNPG services instead of `postgres:5432`.
- Old PostgreSQL data was migrated to CNPG during the initial setup.

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
