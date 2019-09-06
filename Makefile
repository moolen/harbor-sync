
version ?= test
IMAGE_REPO = quay.io/moolen/harbor-sync
IMG ?= ${IMAGE_REPO}:${version}
CRD_OPTIONS ?= "crd:trivialVersions=true"

GOPATH=$(shell go env GOPATH)
HUGO=bin/hugo
KUBECTL=bin/kubectl
MISSPELL=bin/misspell
CONTROLLER_GEN=bin/controller-gen

all: controller

# Run tests
test: generate fmt vet manifests misspell
	go test ./... -coverprofile cover.out

.PHONY: docs
docs: bin/hugo
	cd docs_src; ../$(HUGO) --theme book --destination ../docs

docs-live: bin/hugo
	cd docs_src; ../$(HUGO) server --minify --theme book

# Build harbor-sync-controller binary
controller: generate fmt vet
	go build -o bin/harbor-sync-controller ./main.go

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet manifests
	go run ./main.go

# Install CRDs into a cluster
install: kubectl-bin manifests
	kustomize build config/crd | $(KUBECTL) apply -f -

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: kubectl-bin manifests
	cd config/manager && kustomize edit set image controller=${IMG}
	kustomize build config/default | $(KUBECTL) apply -f -

# Checks if generated files differ
check-gen-files: docs quick-install
	git diff --exit-code

# Generate manifests e.g. CRD, RBAC etc.
manifests: bin/controller-gen
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=harbor-sync paths="./..." output:crd:artifacts:config=config/crd/bases

quick-install: bin/kubectl
	$(KUBECTL) kustomize config/default/ > install/kubernetes/quick-install.yaml

misspell: bin/misspell
	$(MISSPELL) \
		-locale US \
		-error \
		api/* pkg/* docs_src/content/* config/* hack/* README.md CONTRIBUTING.md

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

# Generate code
generate: bin/controller-gen
	$(CONTROLLER_GEN) object:headerFile=./hack/boilerplate.go.txt paths="./..."

# Run tests in container
docker-test:
	rm -rf bin
	docker build -t test:latest -f Dockerfile.test .
	docker run test:latest

# Build the docker image
docker-build:
	docker build . -t ${IMG}

# Push the docker image
docker-push:
	docker push ${IMG}

docker-push-latest:
	docker tag ${IMG} ${IMAGE_REPO}:latest
	docker push ${IMAGE_REPO}:latest

docker-release: docker-build docker-push docker-push-latest

release: quick-install controller docker-release
	tar cvzf bin/harbor-sync-controller.tar.gz bin/harbor-sync-controller

bin/misspell:
	curl -sL https://github.com/client9/misspell/releases/download/v0.3.4/misspell_0.3.4_linux_64bit.tar.gz | tar -xz -C /tmp/
	mkdir bin; cp /tmp/misspell bin/misspell

bin/hugo:
	curl -sL https://github.com/gohugoio/hugo/releases/download/v0.57.2/hugo_extended_0.57.2_Linux-64bit.tar.gz | tar -xz -C /tmp/
	mkdir bin; cp /tmp/hugo bin/hugo

bin/kubectl:
	curl -LO https://storage.googleapis.com/kubernetes-release/release/v1.15.0/bin/linux/amd64/kubectl
	chmod +x ./kubectl
	mkdir bin; mv kubectl bin/kubectl

# find or download controller-gen
# download controller-gen if necessary
bin/controller-gen:
	go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.2.0
	mkdir bin; mv $(GOPATH)/bin/controller-gen bin/controller-gen
