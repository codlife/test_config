## Steps

### 0. Prerequisite

- disable selinux
- stop firewalld
- command: swapoff -a
- command: echo "1" > /proc/sys/net/bridge/bridge-nf-call-iptables


### 1. install docker

| Name       | Version |  Packages  |   
| ------     | ------  | ------ |
| Docker     | 18.09   | [redhat](https://docs.docker.com/install/linux/docker-ee/rhel/), [openSUSE/SUSE](https://docs.docker.com/install/linux/docker-ee/suse/), [centos](https://docs.docker.com/install/linux/docker-ce/centos/), [debian](https://docs.docker.com/install/linux/docker-ce/debian/), [fedora](https://docs.docker.com/install/linux/docker-ce/fedora/), [ubuntu](https://docs.docker.com/install/linux/docker-ce/ubuntu/) |

### 2. install kubernetes

| Name       | Version |  Packages  |   
| ------     | ------  | ------ |
| Kubernetes | 1.13.3  | [redhat/CentOS/SUSE/openSUSE](https://github.com/kubesys/kube-os/releases/download/1.0/kube-tools-v1.13.3-cloudplus.1903.x86_64.rpm), [ubuntu/debian](https://github.com/kubesys/kube-os/releases/download/1.0/kube-tools-v1.13.3-cloudplus.1903.amd64.deb) |

Next, you can use [kubeadm to install kubernetes](https://kubernetes.io/docs/setup/independent/create-cluster-kubeadm/)

Then, you should install a network plugin for kubernetes

```
kubectl create -f https://raw.githubusercontent.com/kubesys/kube-syspods/master/kube-networks/calico/yamls/etcd.yaml 
kubectl create -f https://raw.githubusercontent.com/kubesys/kube-syspods/master/kube-networks/calico/yamls/calico.yaml
```
### 3. install helm

patch 1: upadte hub's docker using the scripts at https://github.com/kubesys/kube-dataAnalysis/tree/master/jupyter/hub

```
docker build . -t jupyterhub/k8s-hub:0.8.0
```

patch 2: upadte singleuser's docker using the scripts at https://github.com/kubesys/kube-dataAnalysis/tree/master/jupyter/singleuser-sample

```
docker build . -t jupyterhub/k8s-singleuser-sample:0.8.0
```

Then, install helm by using the following scripts.

```
  cp ../helm/helm /usr/bin/
  kubectl --namespace kube-system create serviceaccount tiller
  kubectl create clusterrolebinding tiller --clusterrole cluster-admin --serviceaccount=kube-system:tiller
  helm init --service-account tiller --wait
  kubectl patch deployment tiller-deploy --namespace=kube-system --type=json --patch='[{"op": "add", "path": "/spec/template/spec/containers/0/command", "value": ["/tiller", "--listen=localhost:44134"]}]'
```

### 4. install Jupyterhub

```
  bash dev.sh
```

patch 3: support kubernetes 1.13

```
kubectl create -f https://raw.githubusercontent.com/kubesys/kube-dataAnalysis/master/jupyter/yamls/admin-pv.yaml
kubectl create -f https://github.com/kubesys/kube-dataAnalysis/blob/master/jupyter/yamls/pv.yaml
```

### Finally 

Execute the command

```
kubectl describe service proxy-public --namespace jhub
```

Then can see the following outputs:

```
Name:                     proxy-public
Namespace:                jhub
Labels:                   app=jupyterhub
                          chart=jupyterhub-0.8.0
                          component=proxy-public
                          heritage=Tiller
                          release=jhub
Annotations:              <none>
Selector:                 component=proxy,release=jhub
Type:                     LoadBalancer
IP:                       10.111.253.110
Port:                     http  80/TCP
TargetPort:               8000/TCP
NodePort:                 http  31589/TCP
Endpoints:                192.168.66.248:8000
Port:                     https  443/TCP
TargetPort:               443/TCP
NodePort:                 https  31890/TCP
Endpoints:                192.168.66.248:443
Session Affinity:         None
External Traffic Policy:  Cluster
Events:                   <none>
```

Now you can use hub at http://192.168.66.248:8000 (admin/admin)

** Note that if you encouner a permission error, please check your disk permissoins **

## BTW: create user

```
curl -X POST "http://10.109.54.72:8081/hub/api/users/henry" -H  "accept: application/json"
```
