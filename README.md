# godemo

Simple Go guestbook web app, deployed to k3s via [Rancher Fleet](https://fleet.rancher.io/) GitOps.

- App: form (name + message) → PostgreSQL, `/healthz` checks DB
- Postgres in the same namespace (`fleet/godemo/postgres.yaml`, PVC 1Gi)
- Image: `ghcr.io/alexuresp/godemo`
- Ingress: `http://godemo.192-168-1-103.traefik.me`

## Layout

```
├── godemo/                  # Go web app
│   ├── Dockerfile
│   ├── go.mod
│   ├── main.go
│   ├── templates/
│   └── .dockerignore
├── fleet/godemo/            # Fleet bundle (Deployment, Service, Ingress, Postgres)
│   ├── fleet.yaml
│   ├── deployment.yaml
│   ├── service.yaml
│   ├── ingress.yaml
│   └── postgres.yaml
├── node-api/                # (future) Node.js + Express
├── spring-app/              # (future) Spring Boot
├── laravel-app/             # (future) Laravel
├── .github/workflows/       # build & push per app
└── examples/gitrepo.yaml    # one-time GitRepo for Continuous Delivery
```

## Prerequisites

- k3s + Rancher (Fleet / Continuous Delivery)
- Public GitHub repo: https://github.com/Alexuresp/godemo
- `kubectl` and a kubeconfig (local file `config` is gitignored)

## Deploy flow

1. **Push to `main`** — GitHub Actions builds and pushes `ghcr.io/alexuresp/godemo:latest`.
2. **Make the GHCR package public** (first time): GitHub → Packages → `godemo` → Package settings → Change visibility → Public. Without this, nodes get `ImagePullBackOff`.
3. **Register the GitRepo once:**

```bash
kubectl --kubeconfig config apply -f examples/gitrepo.yaml
```

Or in Rancher: **Continuous Delivery** → create Git Repo → URL `https://github.com/Alexuresp/godemo`, branch `main`, path `fleet`, target the local cluster.

4. **Check:**

```bash
kubectl --kubeconfig config -n fleet-local get gitrepo godemo
kubectl --kubeconfig config -n godemo get deploy,svc,ingress,pods
```

5. Open http://godemo.192-168-1-103.traefik.me

## Local run (optional)

```bash
cd godemo && go run .
# http://localhost:8080
```

```bash
cd godemo && docker build -t godemo:local .
docker run --rm -p 8080:8080 godemo:local
```

## Verify Fleet is present

```bash
kubectl --kubeconfig config get ns fleet-local
```
