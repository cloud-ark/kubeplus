FROM golang:1.20.1-bullseye as builder
RUN mkdir /build
ADD ./ /build/
WORKDIR /build/platform-operator
RUN eexport GOOS=linux; go build .

FROM fedora
COPY --from=builder /build/platform-operator/platform-operator /
ENTRYPOINT ["/platform-operator"]
