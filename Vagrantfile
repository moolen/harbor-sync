# -*- mode: ruby -*-
# vi: set ft=ruby :
# based on: https://gist.github.com/avthart/08c5bbdc883ea8e0817141577b4f12fe

$script = <<-SCRIPT
set -x
apt-get update
apt-get install -y docker.io python3 python3-pip
pip3 install --upgrade pip
pip3 install docker-compose --upgrade
curl -sL https://github.com/goharbor/harbor/releases/download/v1.10.1/harbor-online-installer-v1.10.1.tgz | tar zxv
cd harbor
export IPADDR=`ifconfig enp0s8 | grep Mask | awk '{print $2}'| cut -f2 -d:`
sed -i "s/^hostname: .*$/hostname: ${IPADDR}.xip.io/g" harbor.yml
sed -i.bak -e '12,18d' harbor.yml

./prepare
./install.sh
echo "access harbor here: http://${IPADDR}.xip.io."
SCRIPT

Vagrant.configure("2") do |config|
  config.vm.box = "ubuntu/xenial64"
  config.vm.hostname = "harbor"
  config.vm.network "private_network", type: "dhcp"
  config.vm.provision "shell", inline: $script
end
