FROM ubuntu:22.04
ARG TARGETARCH
COPY crd-hook-${TARGETARCH} /crd-hook
ENTRYPOINT ["/crd-hook"]