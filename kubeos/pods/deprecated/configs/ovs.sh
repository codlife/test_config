##################################################
##
## Copyright (2019, ) Institute of Software
##   Chinese Academy of Sciences
##    wuheng@otcaix.iscas.ac.cn
##
##################################################


nic=$(ls /etc/sysconfig/network-scripts/ | grep ifcfg-$1)
if [[ -z $nic ]]
then
  echo "you may input a wrong nic name"
  exit
fi


cp /etc/sysconfig/network-scripts/ifcfg-$1 ifcfg-$1-bak

\cp ifcfg-$1-bak /etc/sysconfig/network-scripts/ifcfg-ovs0
sed -i "s/$1/ovs0/g" /etc/sysconfig/network-scripts/ifcfg-ovs0
sed -i '/DEVICETYPE/d' /etc/sysconfig/network-scripts/ifcfg-ovs0
sed -i '/OVSBOOTPROTO/d' /etc/sysconfig/network-scripts/ifcfg-ovs0
sed -i '/TYPE=/d' /etc/sysconfig/network-scripts/ifcfg-ovs0
sed -i '/UUID/d' /etc/sysconfig/network-scripts/ifcfg-ovs0
echo "DEVICETYPE=ovs" >> /etc/sysconfig/network-scripts/ifcfg-ovs0
echo "OVSBOOTPROTO=none" >> /etc/sysconfig/network-scripts/ifcfg-ovs0
echo "TYPE=OVSBridge" >> /etc/sysconfig/network-scripts/ifcfg-ovs0

\cp ifcfg-NIC.template /etc/sysconfig/network-scripts/ifcfg-$1
sed -i "s/NIC/$1/g" /etc/sysconfig/network-scripts/ifcfg-$1

ovs-vsctl del-br ovs0
ovs-vsctl add-br ovs0
ovs-vsctl add-port ovs0 $1
systemctl restart network
