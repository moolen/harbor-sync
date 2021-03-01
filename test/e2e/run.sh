#!/bin/bash

# Copyright 2019 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
set -o errexit
set -o nounset
set -o pipefail

if ! command -v kind --version &> /dev/null; then
  echo "kind is not installed. Use the package manager or visit the official site https://kind.sigs.k8s.io/"
  exit 1
fi

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd $DIR
export K8S_VERSION=${K8S_VERSION:-v1.20.2}

echo "building container"
make -C ${DIR}/../../ docker-build IMG=harbor-sync:dev
make -C ${DIR} e2e-image IMG=harbor-sync-e2e:dev

echo "copying docker images to cluster..."
kind load docker-image --name="${KIND_CLUSTER_NAME}" harbor-sync:dev
kind load docker-image --name="${KIND_CLUSTER_NAME}" harbor-sync-e2e:dev

echo -e "Granting permissions to e2e service account..."
kubectl create serviceaccount harbor-sync-e2e || true
kubectl create clusterrolebinding permissive-binding \
  --clusterrole=cluster-admin \
  --user=admin \
  --user=kubelet \
  --serviceaccount=default:harbor-sync-e2e || true

echo -e "Waiting service account..."; \
until kubectl get secret | grep -q -e ^harbor-sync-e2e-token; do \
  echo -e "waiting for api token"; \
  sleep 3; \
done

echo -e "Starting the e2e test pod"
FOCUS=${FOCUS:-.*}
export FOCUS

HARBOR_VERSION=${HARBOR_VERSION:-v2.2.0}

helm repo add harbor https://helm.goharbor.io || true
helm upgrade \
  --wait \
  --install harbor harbor/harbor \
  --set nginx.image.tag=${HARBOR_VERSION} \
  --set portal.image.tag=${HARBOR_VERSION} \
  --set core.image.tag=${HARBOR_VERSION} \
  --set jobservice.image.tag=${HARBOR_VERSION} \
  --set registry.image.tag=${HARBOR_VERSION} \
  --set database.internal.image.tag=${HARBOR_VERSION} \
  --set redis.internal.image.tag=${HARBOR_VERSION} \
  -f ./helm-values.yaml

kubectl run --rm \
  --attach \
  --restart=Never \
  --generator=run-pod/v1 \
  --env="FOCUS=${FOCUS}" \
  --overrides='{ "apiVersion": "v1", "spec":{"serviceAccountName": "harbor-sync-e2e"}}' \
  e2e --image=harbor-sync-e2e:dev
