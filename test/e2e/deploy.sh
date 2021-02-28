#!/bin/bash
export NAMESPACE=$1

cat <<EOF > ./kustomization.yaml
namespace: ${NAMESPACE}
bases:
- ./k8s/overlay
EOF

kustomize build . | kubectl apply -f -
sleep 1
kubectl apply -n ${NAMESPACE} -f ./k8s/overlay/sync-config.yaml
kubectl -n ${NAMESPACE} rollout status -w deployment/harbor-sync
