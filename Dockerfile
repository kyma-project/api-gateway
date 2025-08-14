# Build the manager binary
FROM --platform=$BUILDPLATFORM golang:1.25.0-alpine AS builder
ARG TARGETOS
ARG TARGETARCH
ARG VERSION

WORKDIR /api-gateway-build
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY main.go main.go
COPY apis/ apis/
COPY controllers/ controllers/
COPY internal/ internal/
COPY manifests/ manifests/

# Build
# the GOARCH has not a default value to allow the binary be built according to the host where the command
# was called. For example, if we call make docker-build in a local env which has the Apple Silicon M1 SO
# the docker BUILDPLATFORM arg will be linux/arm64 when for Apple x86 it will be linux/amd64. Therefore,
# by leaving it empty we can ensure that the container and binary shipped on it will have the same platform.
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -ldflags="-s -w -X 'github.com/kyma-project/api-gateway/internal/version.version=${VERSION:-}'" -o manager main.go


FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /
COPY --from=builder /api-gateway-build/manager .
COPY --from=builder /api-gateway-build/manifests/ manifests

USER 65532:65532

ENTRYPOINT ["/manager"]
