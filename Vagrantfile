# -*- mode: ruby -*-
# vi: set ft=ruby :

# All Vagrant configuration is done below. The "2" in Vagrant.configure
# configures the configuration version (we support older styles for
# backwards compatibility). Please don't change it unless you know what
# you're doing.
Vagrant.configure("2") do |config|
  # The most common configuration options are documented and commented below.
  # For a complete reference, please see the online documentation at
  # https://docs.vagrantup.com.

  # Every Vagrant development environment requires a box. You can search for
  # boxes at https://vagrantcloud.com/search.
  config.vm.box = "bento/ubuntu-18.04"

  # Disable automatic box update checking. If you disable this, then
  # boxes will only be checked for updates when the user runs
  # `vagrant box outdated`. This is not recommended.
  # config.vm.box_check_update = false

  # Create a forwarded port mapping which allows access to a specific port
  # within the machine from a port on the host machine. In the example below,
  # accessing "localhost:8080" will access port 80 on the guest machine.
  # NOTE: This will enable public access to the opened port
  # config.vm.network "forwarded_port", guest: 80, host: 8080

  # Create a forwarded port mapping which allows access to a specific port
  # within the machine from a port on the host machine and only allow access
  # via 127.0.0.1 to disable public access
  # config.vm.network "forwarded_port", guest: 80, host: 8080, host_ip: "127.0.0.1"

  # Create a private network, which allows host-only access to the machine
  # using a specific IP.
  # config.vm.network "private_network", ip: "192.168.33.10"

  # Create a public network, which generally matched to bridged network.
  # Bridged networks make the machine appear as another physical device on
  # your network.
  # config.vm.network "public_network"

  # Share an additional folder to the guest VM. The first argument is
  # the path on the host to the actual folder. The second argument is
  # the path on the guest to mount the folder. And the optional third
  # argument is a set of non-required options.
  # config.vm.synced_folder "../data", "/vagrant_data"

  # Provider-specific configuration so you can fine-tune various
  # backing providers for Vagrant. These expose provider-specific options.
  # Example for VirtualBox:
  #
  config.vm.provider "virtualbox" do |vb|
  #   # Display the VirtualBox GUI when booting the machine
  #   vb.gui = true
  #
  #   # Customize the amount of memory on the VM:
  #   vb.memory = "1024"
      vb.customize ["modifyvm", :id, "--cpus", 2]
      vb.customize ["modifyvm", :id, "--memory", "3072"]
  end
  #
  # View the documentation for the provider you are using for more
  # information on available options.

  # Enable provisioning with a shell script. Additional provisioners such as
  # Ansible, Chef, Docker, Puppet and Salt are also available. Please see the
  # documentation for more information about their specific syntax and use.
  config.vm.provision "shell", inline: <<-SHELL
     apt-get update
     echo "deb [signed-by=/usr/share/keyrings/cloud.google.gpg] https://packages.cloud.google.com/apt cloud-sdk main" | sudo tee -a /etc/apt/sources.list.d/google-cloud-sdk.list
     sudo apt-get install -y apt-transport-https ca-certificates gnupg
     sudo curl https://packages.cloud.google.com/apt/doc/apt-key.gpg | sudo apt-key --keyring /usr/share/keyrings/cloud.google.gpg add -
     sudo apt-get update && sudo apt-get install -y google-cloud-sdk  apt-transport-https     ca-certificates     curl     gnupg     lsb-release
     sudo curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg
     echo "deb [arch=amd64 signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
     sudo apt-get update
     sudo apt-get install -y docker-ce docker-ce-cli containerd.io
     #sudo groupadd docker
     sudo usermod -aG docker $USER
     sudo curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
     sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl

     sudo curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64
     sudo install minikube-linux-amd64 /usr/local/bin/minikube

     sudo curl https://get.helm.sh/helm-v3.6.3-linux-amd64.tar.gz
     sudo tar -zxvf helm-v3.6.3-linux-amd64.tar.gz
     sudo mv linux-amd64/helm /usr/local/bin/helm

     sudo apt-get install -y python3-pip jq

     pip3 install -r /vagrant/grapher/requirements.txt
     
  SHELL
end


#mv /vagrant/.kube/config /vagrant/.kube/config.orig
#sed 's/C:.*\.minikube/\/vagrant\/.minikube/'g /vagrant/.kube/config.orig | sed 's/\\/\//'g > /vagrant/.kube/config
#sed 's/C:.*\.minikube/\/vagrant\/.minikube/'g config.orig | sed 's/\\/\//'g > config
# cp -r /vagrant/.kube ~/.
#kubectl connections Pod etcd-minikube kube-system --kubeconfig=/vagrant/.kube/config -o png --ignore=ServiceAccount:default,Namespace:default
# cp /home/vagrant/plugins/connections-op.json.gv.png /vagrant/.

# Pre-req
# Install VirtualBox and Minikube on your Windows machine (latest should be fine)
# Start the minikube cluster (minikube start)

# Install latest Vagrant (latest is fine)
# On your Windows machine, open command prompt (cmd) and save the Vagrantfile available in the KubePlus repository in your Windows User home directory.
# Keep its name as Vagrantfile. Then run following commands:
# vagrant up (this will create a Ubuntu Virtual Machine on your Windows Machine. It will take few minutes.)
# vagrant ssh (once the Ubuntu machine is up, ssh into it)
# sudo usermod -aG docker $USER (Add the 'vagrant' user to the docker group so that docker commands can be run without 'sudo')
# exit (for above command to take effect, we have to exit and log back in)
# vagrant ssh 
# docker ps (verify that you are able to run docker commands without sudo)
# wget https://github.com/cloud-ark/kubeplus/raw/master/scripts/pre-kube-env-windows.sh
# chmod +x prep-kube-env-windows.sh
# ./prep-kube-env-windows.sh
# Download and install KubePlus connections plugin
# kubectl connections Pod etcd-minikube kube-system -o png (check that plugin works correctly)
# cp plugins/connections-op.json.gv.png /vagrant/.
# From your Windows machine navigate to the folder where you stored the Vagrantfile. The png file from previous command will be available there.
