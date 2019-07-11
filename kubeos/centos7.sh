############################################
##
## Copyright (2019, ) Institute of Software
##      Chinese Academy of Sciences
##       wuheng@otcaix.iscas.ac.cn
##
############################################

## download repo
curl https://raw.githubusercontent.com/kubesys/kubeos/master/repo/CentOS-OpenStack-rocky.repo > /etc/yum.repos.d/CentOS-OpenStack-rocky.repo
curl https://raw.githubusercontent.com/kubesys/kubeos/master/repo/CentOS-QEMU-EV.repo > /etc/yum.repos.d/CentOS-QEMU-EV.repo
curl https://raw.githubusercontent.com/kubesys/kubeos/master/repo/docker-ce.repo >  /etc/yum.repos.d/docker-ce.repo
curl https://raw.githubusercontent.com/kubesys/kubeos/master/repo/kubernetes.repo > /etc/yum.repos.d/kubernetes.repo

## disable selinux and firewalld
curl https://raw.githubusercontent.com/kubesys/kubeos/master/configs/selinux > /etc/selinux/config
systemctl stop firewalld
systemctl disable firewalld

## install and config software
yum install docker-ce qemu-kvm qemu-img openvswitch -y
systemctl start docker 
systemctl enable docker
systemctl start openvswitch 
systemctl enable openvswitch
rpm --force -Uvh https://github.com/kubesys/kube-os/releases/download/1.1/kube-tools-v1.13.3-cloudplus.1903.x86_64.rpm

## install kubernetes and its network plugin
curl https://raw.githubusercontent.com/kubesys/kubeos/master/syspods/kubernetes/images.conf > images.conf
curl https://raw.githubusercontent.com/kubesys/kubeos/master/syspods/kubernetes/pull.sh | sh
curl https://raw.githubusercontent.com/kubesys/kubeos/master/syspods/kubernetes/dev.sh | sh

while true
  do
    stat=`kubectl get po --all-namespaces | grep kube-apiserver | awk '{print $4}'`
    if [ "$stat" = "Running" ]
    then
      break
    fi
    sleep 5
done

kubectl create -f https://raw.githubusercontent.com/kubesys/kubeos/master/syspods/kube-networks/calico/yamls/etcd.yaml
kubectl create -f https://raw.githubusercontent.com/kubesys/kubeos/master/syspods/kube-networks/calico/yamls/calico.yaml
