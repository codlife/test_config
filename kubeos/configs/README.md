## 1. Manual

If you want to deploy the above softwares on CentOS 7, you can follow the steps.

### 1.1 Prerequisite


1.1.1 disable selinux (vi /etc/selinux/config)

```
SELINUX=(enforcing --> disabled)
```

1.1.2 copy all the *.repo you needs to the path ``/etc/yum.repos.d/''

1.1.3 disable firewalld
```
systemctl stop firewalld
systemctl disable firewalld
```

### 1.2 Install Docker

```
yum install docker-ce
systemctl start docker 
systemctl enable docker
```

### 1.3 Install Kubernetes

```
yum install kubeadm kubectl kubelet  
or 
rpm --force -Uvh https://github.com/kubesys/kube-os/releases/download/1.0/kube-tools-v1.13.3-cloudplus.1903.x86_64.rpm
```

### 1.4 Install kvm （optional）

```
yum install qemu-kvm qemu-img
```

### 1.5 Install openvswitch （optional）

```
yum install openvswitch
systemctl start openvswitch 
systemctl enable openvswitch
```

Next, please see project [syspods](../syspods) to complete installation.
