VERSION := $(shell tail -n 1 versions.txt)

.PHONY: latest versioned

latest:
	docker build -t gcr.io/cloudark-kubeplus/grapher:latest .

versioned:
	docker build -t gcr.io/cloudark-kubeplus/grapher:$(VERSION) .
	docker push gcr.io/cloudark-kubeplus/grapher:$(VERSION)