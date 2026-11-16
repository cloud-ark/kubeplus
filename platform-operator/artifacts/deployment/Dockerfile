FROM fedora
ARG TARGETARCH
ADD platform-operator-${TARGETARCH} /platform-operator
ENTRYPOINT ["/platform-operator"]