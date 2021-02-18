# Build the manager binary
FROM alpine

WORKDIR /
# Copy the Go Modules manifests
COPY bin/manager .

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
# USER nonroot:nonroot

ENTRYPOINT ["/manager"]