# Build the manager binary
FROM golang:1.23.2-alpine AS builder
ARG TARGET_OS
ARG TARGET_ARCH
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
RUN CGO_ENABLED=0 GOOS=${TARGET_OS:-linux} GOARCH=${TARGET_ARCH:-amd64} go build -ldflags="-X 'github.com/kyma-project/api-gateway/internal/version.version=${VERSION:-}'" -a -o manager main.go


# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /api-gateway-build/manager .
COPY --from=builder /api-gateway-build/manifests/ manifests

USER 65532:65532

ENTRYPOINT ["/manager"]
