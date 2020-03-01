FROM golang:1.14 as BASE
RUN go get -u github.com/onsi/ginkgo/ginkgo
RUN curl -LO https://storage.googleapis.com/kubernetes-release/release/v1.17.3/bin/linux/amd64/kubectl && \
    chmod +x ./kubectl && \
    mv kubectl /usr/local/bin/kubectl

RUN curl -LO https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize%2Fv3.5.4/kustomize_v3.5.4_linux_amd64.tar.gz && \
    tar xvf kustomize* && \
    chmod +x ./kustomize && \
    mv ./kustomize /usr/local/bin/kustomize


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

COPY entrypoint.sh      /entrypoint.sh
COPY e2e.test           /e2e.test
COPY deploy.sh          /deploy.sh
COPY k8s                /k8s

CMD [ "/entrypoint.sh" ]
