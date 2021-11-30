FROM golang:1.16 as BASE
RUN go get -u github.com/onsi/ginkgo/ginkgo
RUN curl -LO https://storage.googleapis.com/kubernetes-release/release/v1.21.2/bin/linux/amd64/kubectl && \
    chmod +x ./kubectl && \
    mv kubectl /usr/local/bin/kubectl

RUN curl -LO https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize%2Fv3.5.4/kustomize_v3.5.4_linux_amd64.tar.gz && \
    tar xvf kustomize* && \
    chmod +x ./kustomize && \
    mv ./kustomize /usr/local/bin/kustomize


RUN curl -LO https://github.com/google/go-containerregistry/releases/download/v0.4.1/go-containerregistry_Linux_x86_64.tar.gz && \
    tar xvf go-containerregistry* && \
    chmod +x ./crane && \
    mv ./crane /usr/local/bin/crane

FROM alpine:3.11
RUN apk add -U --no-cache \
    ca-certificates \
    bash \
    curl \
    tzdata \
    libc6-compat \
    openssl

COPY --from=BASE /go/bin/ginkgo /usr/local/bin/
COPY --from=BASE /usr/local/bin/kubectl /usr/local/bin/
COPY --from=BASE /usr/local/bin/kustomize /usr/local/bin/
COPY --from=BASE /usr/local/bin/crane /usr/local/bin/

# add harbor ca cert
# this was extracted from harbor-secret.yaml
ADD ca.pem /usr/local/share/ca-certificates
RUN update-ca-certificates

COPY entrypoint.sh      /entrypoint.sh
COPY e2e.test           /e2e.test
COPY deploy.sh          /deploy.sh
COPY k8s                /k8s

CMD [ "/entrypoint.sh" ]
