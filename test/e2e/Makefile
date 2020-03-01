IMG=harbor-sync-e2e:dev

.PHONY: e2e-bin
e2e-bin:
	ginkgo build .

.PHONY: e2e-image
e2e-image: e2e-bin
	mkdir -p k8s
	-rm -rf k8s/base
	cp -r ../../config k8s/base
	docker build -t ${IMG} .
