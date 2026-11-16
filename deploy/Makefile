VERSION := $(shell tail -n 1 versions.txt)

.PHONY: latest versioned

latest:
	docker build -t gcr.io/cloudark-kubeplus/webhook-tls-getter:latest .

versioned:
	docker build -t gcr.io/cloudark-kubeplus/webhook-tls-getter:$(VERSION) .
	docker push gcr.io/cloudark-kubeplus/webhook-tls-getter:$(VERSION)