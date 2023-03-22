#FROM alpine:latest
FROM golang:1.20.1-bullseye as builder
RUN mkdir /build
ADD ./ /build/
WORKDIR /build/mutating-webhook
RUN  GO111MODULE=on; CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o crd-hook

FROM ubuntu:20.04 
COPY --from=builder /build/mutating-webhook/crd-hook crd-hook
ENTRYPOINT ["./crd-hook"]

