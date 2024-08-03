FROM golang:1.20.1-bullseye as builder
RUN mkdir /build
ADD ./ /build/
WORKDIR /build/
RUN export GOOS=linux; go build .

FROM fedora
COPY --from=builder /build/platform-operator /
ENTRYPOINT ["/platform-operator"]
