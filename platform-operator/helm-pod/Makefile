VERSION := $(shell tail -n 1 versions.txt)

.PHONY: latest versioned build

latest: build
	docker build --no-cache -t gcr.io/cloudark-kubeplus/helm-pod:latest .

versioned: build
	docker build -t gcr.io/cloudark-kubeplus/helm-pod:$(VERSION) .
	docker push gcr.io/cloudark-kubeplus/helm-pod:$(VERSION)

build:
	GO111MODULE=on CGO_ENABLED=0 GOOS=linux go build -o helm-pod