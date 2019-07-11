## 1. Background


In our design, the OS includes Docker, kubernetes, and you can get the packages by yourself.
Note that you should first disable selinux and firewalld of your OS.

| Name       | Version |  Packages  |   
| ------     | ------  | ------ |
| Docker     | 18.09   | [redhat](https://docs.docker.com/install/linux/docker-ee/rhel/), [openSUSE/SUSE](https://docs.docker.com/install/linux/docker-ee/suse/), [centos](https://docs.docker.com/install/linux/docker-ce/centos/), [debian](https://docs.docker.com/install/linux/docker-ce/debian/), [fedora](https://docs.docker.com/install/linux/docker-ce/fedora/), [ubuntu](https://docs.docker.com/install/linux/docker-ce/ubuntu/) |
| Kubernetes | 1.13.5  | [redhat/CentOS/SUSE/openSUSE](https://github.com/kubesys/kubeos/releases/download/1.2/kube-tools-v1.13.5-cloudplus.1903.el7.x86_64.rpm), [ubuntu/debian](https://github.com/kubesys/kubeos/releases/download/1.2/kube-tools-v1.13.5-cloudplus.1903.amd64.deb) |
| Helm        | 2.13.1   | [Linux](https://storage.googleapis.com/kubernetes-helm/helm-v2.13.1-linux-amd64.tar.gz) |
| KVM        | 2.12   | [redhat/CentOS](https://docs.openstack.org/install-guide/environment-packages-rdo.html), [openSUSE/SUSE](https://docs.openstack.org/install-guide/environment-packages-obs.html), [debian/ubuntu](https://docs.openstack.org/install-guide/environment-packages-ubuntu.html) |
| openvswitch| 2.10   | [redhat/CentOS](http://docs.openvswitch.org/en/latest/intro/install/distributions/#red-hat), [openSUSE/SUSE](http://docs.openvswitch.org/en/latest/intro/install/distributions/#opensuse), [debian/ubuntu](http://docs.openvswitch.org/en/latest/intro/install/distributions/#debian) |
| vpp        | 19.01  | [Linux](https://wiki.fd.io/view/VPP/Installing_VPP_binaries_from_packages) |

updated: 2019-4-12


## 2. Setup for CentOS7

2.1 install docker, disable selinux and disable firwalld

2.2 download kubeos

```
git clone https://github.com/kubesys/kubeos.git
rpm --force -Uvh https://github.com/kubesys/kubeos/releases/download/1.2/kube-tools-v1.13.5-cloudplus.1903.el7.x86_64.rpm
```

2.3 install kubernetes

```
cd syspods/kubernetes
bash pull.sh
bash dev.sh
```

2.4 install calico

```
cd syspods/calico
bash dev.sh
```

waiting for a few minutes, ensure that all pods are running

```
[root@iscassystems calico]# kubectl get po -n kube-system
NAME                                       READY   STATUS    RESTARTS   AGE
calico-etcd-94d8m                          1/1     Running   0          3m35s
calico-kube-controllers-74887d7bdf-8j8ms   1/1     Running   2          5m42s
calico-node-ms29r                          1/1     Running   4          5m42s
coredns-86c58d9df4-cd5sz                   1/1     Running   0          18m
coredns-86c58d9df4-kzgd7                   1/1     Running   0          18m
etcd-iscassystems                          1/1     Running   0          18m
kube-apiserver-iscassystems                1/1     Running   0          18m
kube-controller-manager-iscassystems       1/1     Running   0          18m
kube-proxy-gv74x                           1/1     Running   0          18m
kube-scheduler-iscassystems                1/1     Running   0          18m
```

2.5 install helm

```
cd syspods/helm
bash dev.sh
```

2.6 install dashboard

```
cd syspods/webui
bash kubedash.sh
```

The outputs:
```
URL: https://<masterip>:32668
token: eyJhbGciOiJSUzI1NiIsImtpZCI6IiJ9.eyJpc3MiOiJrdWJlcm5ldGVzL3NlcnZpY2VhY2NvdW50Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9uYW1lc3BhY2UiOiJrdWJlLXN5c3RlbSIsImt1YmVybmV0ZXMuaW8vc2VydmljZWFjY291bnQvc2VjcmV0Lm5hbWUiOiJrdWJlcm5ldGVzLWRhc2hib2FyZC10b2tlbi05OTQyMiIsImt1YmVybmV0ZXMuaW8vc2VydmljZWFjY291bnQvc2VydmljZS1hY2NvdW50Lm5hbWUiOiJrdWJlcm5ldGVzLWRhc2hib2FyZCIsImt1YmVybmV0ZXMuaW8vc2VydmljZWFjY291bnQvc2VydmljZS1hY2NvdW50LnVpZCI6ImQxNjUwNjNjLTUzNjEtMTFlOS1hNzg0LTAwMGMyOTA0NDI4NyIsInN1YiI6InN5c3RlbTpzZXJ2aWNlYWNjb3VudDprdWJlLXN5c3RlbTprdWJlcm5ldGVzLWRhc2hib2FyZCJ9.eMAcEsWKDhQMbkCog4QPXmkhLRJTpy_E0XnuoxcoSvKcyaCG5aFoq6bbNFmAMco-Wc-VdOnUvdoSwe1YEC67oZBDrcMfa3jtVnxfIBsBoA0Vj-CN-Sw0KMAdvy6qKq1pyh-fdm15lT25TwohF-aPOKX2ybUMYRFBgzB0ao0SQ_kzcDY-nswlifwg3MsNUG9y0MR8S5AAvC422FaJY1P0awLwbMu_WulE2AyjW1YTvUkemZYk8iEZgFTYYiq4veT6OZWm8-FfcL0ic69Stznk0qyWEPBwQn952-22S19PUlE0ZjrCNaVSFEDErjKkx4uBzIXlR6vmAaPL-Z3skXDw3w
```


## 3. List

- [yum repos](configs): yum client for CentOS.7x
- [kubernetes pods](pods): system pods 
