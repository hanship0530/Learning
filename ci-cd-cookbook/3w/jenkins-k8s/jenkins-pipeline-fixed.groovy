cat <<EOF | kubectl apply -f -
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: dev-nginx
  namespace: argocd
  finalizers:
  - resources-finalizer.argocd.argoproj.io
spec:
  project: default
  source:
    helm:
      valueFiles:
      - values-dev.yaml
    path: nginx-chart
    repoURL: https://github.com/hanship0530/ops-deploy
    targetRevision: HEAD
  syncPolicy:
    automated:
      prune: true
    syncOptions:
    - CreateNamespace=true
  destination:
    namespace: dev-nginx
    server: https://kubernetes.default.svc
EOF