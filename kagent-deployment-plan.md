# kagent Deployment Plan for k3sdemo

## 1. Цель
Развернуть **kagent** (Kubernetes-native agent runtime от Solo.io, CNCF Sandbox) в кластере `k3sdemo` для автоматизации операций с Fleet, CNPG, MinIO и приложениями. Агенты живут как CRD + GitOps.

## 2. LLM-провайдер: OpenRouter
- Base URL: `https://openrouter.ai/api/v1`
- Auth: Bearer token
- **Важно**: OpenRouter API key хранится в Kubernetes Secret, НЕ коммитится в Git.
- Доступные модели: OpenAI/Anthropic/Gemini/Mixtral/etc через OpenRouter.
- Текущая модель: `openai/gpt-oss-120b`

## 3. Архитектура в рамках k3sdemo

```
┌─────────────────────────────────────────────────────┐
│                    k3s Cluster                       │
│                                                     │
│  MetalLB 192.168.1.151                              │
│  Traefik Ingress                                    │
│  ┌─────────────────────────────────────────────┐    │
│  │ kagent namespace                             │    │
│  │ ┌───────────────────┐   ┌────────────────┐  │    │
│  │ │ kagent-controller │   │ kagent-ui      │  │    │
│  │ │ (manages CRDs)    │   │ Ingress:       │  │    │
│  │ │                   │   │ kagent.192-168-│  │    │
│  │ └─────────┬─────────┘   │ -1-151.traefik │  │    │
│  │           │             │ .me            │  │    │
│  │  CRDs:   │             └────────────────┘  │    │
│  │  - Agent │                                 │    │
│  │  - Session                                 │    │
│  │  - Tool                                    │    │
│  │  - AgentHarness                            │    │
│  │           │                                 │    │
│  │  Agent:   │                                 │    │
│  │  - cnpg-admin   → управление CNPG          │    │
│  │  - fleet-operator → Fleet/GitOps           │    │
│  │  - k8s-admin     → админка кластера        │    │
│  │  - minio-helper  → работа с бэкапами       │    │
│  │           │                                 │    │
│  │  MCP Tools:                                 │    │
│  │  - kubectl, helm, pg_restore, mc           │    │
│  └─────────────────────────────────────────────┘    │
│                                                     │
│  CNPG: godemo, godemo-dev, laravel                  │
│  MinIO: http://10.132.7.1:9000 bucket cnpg-backups │
│  Fleet GitOps: godemo, godemo-dev, node-api, ...   │
└─────────────────────────────────────────────────────┘
```

## 4. Пошаговый план

### Шаг 1. Установить kagent CLI
```bash
curl https://raw.githubusercontent.com/kagent-dev/kagent/refs/heads/main/scripts/get-kagent | bash
```

### Шаг 2. Namespace + Secret
```bash
kubectl create namespace kagent
kubectl create secret generic openrouter-secret \
  --namespace kagent \
  --from-literal=apiKey=<OPENROUTER_API_KEY>
```

### Шаг 3. Helm install
```bash
helm repo add kagent https://kagent-dev.github.io/kagent/helm
helm repo update

helm install kagent-crds kagent/kagent-crds \
  --namespace kagent --create-namespace --wait

helm install kagent kagent/kagent \
  --namespace kagent \
  --set providers.default=openRouter \
  --set providers.openRouter.baseURL="https://openrouter.ai/api/v1" \
  --set providers.openRouter.apiKey="<OPENROUTER_API_KEY>"
```

### Шаг 4. UI Ingress
```yaml
# fleet/kagent/ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: kagent-ui
  namespace: kagent
  annotations:
    traefik.ingress.kubernetes.io/router.entrypoints: web
spec:
  ingressClassName: traefik
  rules:
    - host: kagent.192-168-1-151.traefik.me
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: kagent-ui
                port:
                  number: 8080
```

### Шаг 5. Agents (CRD)
- `cnpg-admin` — управление CNPG кластерами
- `fleet-operator` — Fleet GitOps операции
- `k8s-admin` — общая админка кластера
- `minio-helper` — работа с MinIO бэкапами

### Шаг 6. MCP Tools (kmcp / custom)
- kubectl, helm, pg_restore, mc
- fleet wrapper для Fleet CRDs

### Шаг 7. Fleet манифесты
- `fleet/kagent/gitrepo.yaml` — если хранить отдельно
- `fleet/kagent/agent-*.yaml` — агенты
- `fleet/kagent/tools.yaml` — MCP tools
- `fleet/kagent/ingress.yaml` — UI

### Шаг 8. Тестирование
- Smoke test: простые вопросы про кластер
- Реальная задача: бэкап CNPG + проверка в MinIO
- Реальная задача: force-sync Fleet + проверка деплоя

## 5. Безопасность
- OpenRouter API key в Secret, не в Git
- RBAC через ServiceAccount + ClusterRole/ClusterRoleBinding
- Опционально: Agent Substrate для sandbox
- Логи: OpenTelemetry + Prometheus + kagent logs

## 6. Rollback
```bash
helm uninstall kagent -n kagent
helm uninstall kagent-crds -n kagent
kubectl delete namespace kagent
```

## 7. Риски
1. OpenRouter rate limits / cost — нужен тариф мониторинг
2. RBAC scope — можно сузить после калибровки
3. Хранение Agent CRD — отдельный GitRepo или в текущем?
4. Доступ к UI — Ingress vs port-forward
5. Agent Substrate — опционально для production

## 8. CLI команды после установки
```bash
kagent get agent
kagent invoke --agent cnpg-admin -t "сделай бэкап godemo"
kagent dashboard
```
