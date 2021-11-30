FROM golang:1.16 as builder

WORKDIR /workspace
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

COPY api/ api/
COPY pkg/ pkg/
COPY cmd/ cmd/
COPY main.go main.go

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o harbor-sync main.go

FROM alpine:3.12
WORKDIR /
RUN apk add --update ca-certificates
COPY --from=builder /workspace/harbor-sync .
