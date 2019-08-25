# -*- mode: ruby -*-
# vi: set ft=ruby :
# based on: https://gist.github.com/avthart/08c5bbdc883ea8e0817141577b4f12fe

$script = <<-SCRIPT
apt-get update
apt-get install -y docker.io python python-pip
pip install docker-compose --upgrade
curl -s https://storage.googleapis.com/harbor-releases/release-1.8.0/harbor-online-installer-v1.8.2.tgz | tar zxv
cd harbor
export IPADDR=`ifconfig enp0s8 | grep Mask | awk '{print $2}'| cut -f2 -d:`
sed -i "s/^hostname: .*$/hostname: ${IPADDR}.xip.io/g" harbor.yml

./prepare
./install.sh
SCRIPT

Vagrant.configure("2") do |config|
  config.vm.box = "ubuntu/xenial64"
  config.vm.hostname = "harbor"
  config.vm.network "private_network", type: "dhcp"
  config.vm.provision "shell", inline: $script
end
