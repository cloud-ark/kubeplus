#!/bin/bash -x

if (( $# < 1 )); then
	echo "Usage: ./kubeplus-control-center.sh <start|stop> <http|https> <domain_name> [<inet_ip>]"
	echo "For Vagrant based dev env:"
	echo "./kubeplus-control-center.sh start http localhost localhost"
	echo "For AWS (e.g.:):"
	echo "./kubeplus-control-center.sh start https kubeplus-saas-manager.com 172.31.6.225"
    exit 0
fi


check_docker() {
  docker ps 2>1 1>/dev/null
  if [[ $? == 1 ]]; then
    echo "Cannot connect to the Docker daemon at unix:///var/run/docker.sock. Is the docker daemon running?."
    echo "Docker is needed to run KubePlus control center"
    exit 0
  fi
}

setup_creds() {

	echo -e "Choose username for control center>\c"
	read -e username

	echo -e "Choose password for control center>\c"
	read -e password

	mkdir -p "$HOME/.kubeplus/"
	echo "$username":"$password" >> "$HOME/.kubeplus/creds"

}

start_control_center() {

	check_docker

	# Setup creds
	if [[ ! -f "$HOME/.kubeplus/creds" ]]; then
		setup_creds
	fi

	# Pull image
	echo "Pulling control center Docker image..."
	docker pull gcr.io/cloudark-kubeplus/kubeplus-saas-manager-control-center:0.37

	minikube_ip=`minikube ip`

	if [[ $? != 0 ]]; then
		minikube_ip=''
	#else
	#	eval $(minikube docker-env)
	fi

	# Start the container
	echo "Starting the container..."
	mkdir -p ~/.kubeplus
	mkdir -p ~/.kubeplus/consoles
	mkdir -p ~/.kubeplus/prometheus-data

	export DOMAIN_NAME=$domain_name

	containerId=`docker run --env MINIKUBE_IP=$minikube_ip --env PROTOCOL=$protocol --env INET_IP=$inet_ip --env DOMAIN_NAME=$domain_name -p 5002:5002 -P --network host -v ~/.kubeplus/:/root/.kube -v ~/.kubeplus/:/root/.kubeplus -v ~/.kubeplus/consoles/:/prometheus-2.27.1.linux-amd64/consoles -v ~/.kubeplus/prometheus-data/:/prometheus-2.27.1.linux-amd64/data  -d gcr.io/cloudark-kubeplus/kubeplus-saas-manager-control-center:0.37`
	echo $containerId >> ~/.kubeplus/monitoring-container.txt
	echo "KubePlus Control Center container: $containerId"

	# Start AlertManager
	docker exec -d $containerId ./alertmanager-0.23.0.linux-amd64/alertmanager --config.file=/alertmanager-0.23.0.linux-amd64/alertmanager.yml --cluster.listen-address=0.0.0.0:65499 --web.listen-address=:65498

	# Start Prometheus
	if [[ $protocol == "https" ]]; then
		docker exec -d $containerId ./prometheus-2.27.1.linux-amd64/prometheus --web.enable-lifecycle --config.file=/root/kubeplus-saas-manager-control-center/prometheus/prometheus.yml --web.listen-address="$inet_ip:65500" --web.console.templates=/prometheus-2.27.1.linux-amd64/consoles --web.console.libraries=/prometheus-2.27.1.linux-amd64/console_libraries --storage.tsdb.path=/prometheus-2.27.1.linux-amd64/data --web.config.file=/root/kubeplus-saas-manager-control-center/prometheus/web-config.yml
	else
		docker exec -d $containerId ./prometheus-2.27.1.linux-amd64/prometheus --web.enable-lifecycle --config.file=/root/kubeplus-saas-manager-control-center/prometheus/prometheus.yml --web.listen-address="0.0.0.0:65500" --web.console.templates=/prometheus-2.27.1.linux-amd64/consoles --web.console.libraries=/prometheus-2.27.1.linux-amd64/console_libraries --storage.tsdb.path=/prometheus-2.27.1.linux-amd64/data
	fi

	# Run the portal
	docker exec -d $containerId /root/kubeplus-saas-manager-control-center/portal/start-portal.sh

	# Run the metrics getter
	docker exec -d $containerId /root/kubeplus-saas-manager-control-center/prometheus/start-metrics-getter.sh

	# Get the control center url
	if [[ $? == 0 ]]; then
		echo "KubePlus Control center started successfully."
		if [ -f /sys/hypervisor/uuid ]; then
    		if [ `head -c 3 /sys/hypervisor/uuid` == "ec2" ]; then
      			ec2ip=`curl http://checkip.amazonaws.com`
      			echo "Access it here: http://$ec2ip:5002/"
    		fi
		else
			if [[ ! -z "$DOCKER_HOST" ]]; then
				#minikubeip=`minikube ip`
				serviceip=`echo $DOCKER_HOST | cut -d ':' -f 2 | cut -d '/' -f 3`
			else
				serviceip="localhost"
			fi;
			echo "Access it at: http://$serviceip:5002"
		fi;
		echo "Use the credentials that you created (available in $HOME/.kubeplus/creds)"
	else
		echo "KubePlus Control center failed to start."
	fi;
}

stop_control_center() {
	check_docker
	container=`tail -1 ~/.kubeplus/monitoring-container.txt`
	echo "Stopping control center container $container."
	op=`docker stop $container`
	op=`docker rm $container`
}

command=$1
protocol=$2
domain_name=$3
inet_ip=$4

if [[ "$command" == "start" ]]; then
	start_control_center $protocol $inet_ip $domain_name
elif [[ "$command" == "stop" ]]; then
	stop_control_center
else
	echo "Unknown option specified."
fi
