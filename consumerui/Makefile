VERSION := $(shell tail -n 1 versions.txt)

.PHONY: latest versioned

latest:
	docker build --no-cache -t gcr.io/cloudark-kubeplus/consumerui:latest .

versioned:
	docker build --no-cache -t gcr.io/cloudark-kubeplus/consumerui:$(VERSION) .
	docker push gcr.io/cloudark-kubeplus/consumerui:$(VERSION)