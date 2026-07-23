# kagent folder

Saved manifests for redeploying kagent in this cluster.

##redeploy
kubectl apply -f fleet/kagent/
kubectl scale deployment --all -n kagent --replicas=1

## Cleanup live pods while keeping config
kubectl scale deployment --all -n kagent --replicas=0

## Verify
kubectl get pods -n kagent
kubectl get modelconfig,agent,secret -n kagent
