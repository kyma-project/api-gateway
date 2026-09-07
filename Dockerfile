# Build the manager binary
FROM --platform=$BUILDPLATFORM golang:1.27.1-alpine AS builder
ARG TARGETOS
ARG TARGETARCH
ARG VERSION

WORKDIR /api-gateway-build
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Copy the go source
COPY apis/ apis/
COPY internal/ internal/
COPY manifests/ manifests/
COPY cmd/ cmd/

# Build
# the GOARCH has no default value to allow the binary to be built according to the host where the command
# was called. For example, if we call make docker-build in a local env which has the Apple Silicon M1 SO
# the docker BUILDPLATFORM arg will be linux/arm64 when for Apple x86 it will be linux/amd64. Therefore,
# by leaving it empty we can ensure that the container and binary shipped on it will have the same platform.
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=$TARGETARCH GOFIPS140=v1.0.0 go build -ldflags="-s -w -X 'github.com/kyma-project/api-gateway/internal/version.version=${VERSION:-}'" -o manager cmd/main.go


FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /api-gateway-build/manager .
COPY --from=builder /api-gateway-build/manifests/ manifests
ENV GODEBUG="fips140=only,tlsmlkem=0"
USER 65532:65532

ENTRYPOINT ["/manager"]
