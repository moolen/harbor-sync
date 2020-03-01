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

KIND_LOG_LEVEL="1"

if ! [ -z $DEBUG ]; then
  set -x
  KIND_LOG_LEVEL="6"
fi

set -o errexit
set -o nounset
set -o pipefail

cleanup() {
  if [[ "${KUBETEST_IN_DOCKER:-}" == "true" ]]; then
    kind "export" logs --name ${KIND_CLUSTER_NAME} "${ARTIFACTS}/logs" || true
  fi

  # kind delete cluster \
  #   --verbosity=${KIND_LOG_LEVEL} \
  #   --name ${KIND_CLUSTER_NAME}
}

trap cleanup EXIT

if ! command -v kind --version &> /dev/null; then
  echo "kind is not installed. Use the package manager or visit the official site https://kind.sigs.k8s.io/"
  exit 1
fi

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd $DIR
export K8S_VERSION=${K8S_VERSION:-v1.17.2@sha256:59df31fc61d1da5f46e8a61ef612fa53d3f9140f82419d1ef1a6b9656c6b737c}
export DOCKER_CLI_EXPERIMENTAL=enabled

KIND_CLUSTER_NAME="harbor-sync-dev"
echo "creating Kubernetes cluster with kind"

export KUBECONFIG="${HOME}/.kube/kind-config-${KIND_CLUSTER_NAME}"
# kind create cluster \
#   --verbosity=${KIND_LOG_LEVEL} \
#   --name ${KIND_CLUSTER_NAME} \
#   --config ${DIR}/kind.yaml \
#   --retain \
#   --image "kindest/node:${K8S_VERSION}"

echo "Kubernetes cluster:"
kubectl get nodes -o wide

echo "building container"

docker build -t fake-harbor-api:dev ${DIR}/../fake-harbor-api
make -C ${DIR}/../../ docker-build IMG=harbor-sync:dev
make -C ${DIR} e2e-image IMG=harbor-sync-e2e:dev

echo "copying docker images to cluster..."
kind load docker-image --name="${KIND_CLUSTER_NAME}" fake-harbor-api:dev
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

kubectl run --rm \
  --attach \
  --restart=Never \
  --generator=run-pod/v1 \
  --env="FOCUS=${FOCUS}" \
  --overrides='{ "apiVersion": "v1", "spec":{"serviceAccountName": "harbor-sync-e2e"}}' \
  e2e --image=harbor-sync-e2e:dev
