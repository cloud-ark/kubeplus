FROM golang:1.20.1-bullseye as builder
RUN mkdir /build
ADD ./ /build/
WORKDIR /build/mutating-webhook-helper
RUN GO111MODULE=on; export GOOS=linux; go build .

FROM ubuntu:20.04
COPY --from=builder /build/mutating-webhook-helper/mutating-webhook-helper mutating-webhook-helper
ENTRYPOINT ["./mutating-webhook-helper"]

