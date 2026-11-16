VERSION := $(shell tail -n 1 versions.txt)

.PHONY: latest versioned build

latest: build
	docker build -t gcr.io/cloudark-kubeplus/platform-operator:latest artifacts/deployment

versioned: build
	docker build -t gcr.io/cloudark-kubeplus/platform-operator:$(VERSION) artifacts/deployment
	docker push gcr.io/cloudark-kubeplus/platform-operator:$(VERSION)

build:
	CGO_ENABLED=0 GOOS=linux go build -o artifacts/deployment/platform-operator