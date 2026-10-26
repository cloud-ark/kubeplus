VERSION := $(shell tail -n 1 versions.txt)

.PHONY: latest versioned build

latest: build
	docker build --no-cache -t gcr.io/cloudark-kubeplus/pac-mutating-admission-webhook:latest .

versioned: build
	docker build -t gcr.io/cloudark-kubeplus/pac-mutating-admission-webhook:$(VERSION) .
	docker push gcr.io/cloudark-kubeplus/pac-mutating-admission-webhook:$(VERSION)

build:
	GO111MODULE=on CGO_ENABLED=0 GOOS=linux go build -installsuffix cgo -o crd-hook