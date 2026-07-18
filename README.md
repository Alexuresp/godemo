# godemo

Simple Go guestbook web app, deployed to k3s via [Rancher Fleet](https://fleet.rancher.io/) GitOps.

- App: form (name + message), in-memory list, `/healthz`
- Image: `ghcr.io/alexuresp/godemo`
- Ingress: `http://godemo.192-168-1-103.traefik.me`

## Layout

```
├── main.go / templates/     # Go web server
├── Dockerfile
├── .github/workflows/       # build & push to GHCR
├── fleet/                   # Fleet bundle (Deployment, Service, Ingress)
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
go run .
# http://localhost:8080
```

```bash
docker build -t godemo:local .
docker run --rm -p 8080:8080 godemo:local
```

## Verify Fleet is present

```bash
kubectl --kubeconfig config get ns fleet-local
```
