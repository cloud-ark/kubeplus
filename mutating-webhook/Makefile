 # Copyright 2017 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
all: docker

docker:
	# dep ensure -v
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o crd-hook
	docker build --no-cache -t gcr.io/${PROJECT_ID}/${IMAGE_NAME}:1.0 .
	rm -rf crd-hook
	docker push gcr.io/${PROJECT_ID}/${IMAGE_NAME}:1.0

deploy:
	kubectl create -f deployment/deployment.yaml
	kubectl create -f deployment/service.yaml
	cat ./deployment/mutatingwebhook.yaml | ./deployment/webhook-patch-ca-bundle.sh > ./deployment/mutatingwebhook-ca-bundle.yaml
	kubectl create -f deployment/mutatingwebhook-ca-bundle.yaml
delete:
	kubectl delete -f deployment/deployment.yaml
	kubectl delete -f deployment/service.yaml
	kubectl delete -f deployment/mutatingwebhook-ca-bundle.yaml
install-operators:
	helm install https://github.com/cloud-ark/operatorcharts/blob/master/mysql-operator-0.2.5-1.tgz?raw=true
	helm install https://github.com/cloud-ark/operatorcharts/blob/master/moodle-operator-chart-0.3.0.tgz?raw=true
	helm install https://github.com/cloud-ark/operatorcharts/blob/master/stash-operator-chart-0.8.4.tgz?raw=true
cluster:
	kubectl create -f deployment/crds/secrets/cluster1-secret.yaml
	kubectl create -f deployment/crds/cluster1.yaml
delcluster:
	kubectl delete -f deployment/crds/secrets/cluster1-secret.yaml
	kubectl delete -f deployment/crds/cluster1.yaml
moodle:
	kubectl create -f deployment/crds/moodle1.yaml
delmoodle:
	kubectl delete -f deployment/crds/moodle1.yaml
restic:
	kubectl create -f deployment/crds/restic-moodle.yaml
delrestic:
	kubectl delete -f deployment/crds/restic-moodle.yaml

delall:
	kubectl delete -f deployment/crds/secrets/cluster1-secret.yaml
	kubectl delete -f deployment/crds/cluster1.yaml
	kubectl delete -f deployment/crds/moodle1.yaml
	kubectl delete -f deployment/crds/restic-moodle.yaml

gen-certs:
	bash ./deployment/webhook-create-signed-cert.sh	--service crd-hook-service --namespace default --secret webhook-tls-certificates
	cat ./deployment/mutatingwebhook.yaml | ./deployment/webhook-patch-ca-bundle.sh > ./deployment/mutatingwebhook-ca-bundle.yaml
clean:
	rm crdhook
