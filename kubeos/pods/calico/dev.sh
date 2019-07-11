############################################
##
## Copyright (2019, ) Institute of Software
##      Chinese Academy of Sciences
##       wuheng@otcaix.iscas.ac.cn
##
############################################

function setupCalico()
{
  while true
  do
    stat=`kubectl get po --all-namespaces | grep kube-apiserver | awk '{print $4}'`
    if [ "$stat" = "Running" ]
    then
      break
    fi
    sleep 5
  done
  
  kubectl create -f yamls/etcd.yaml
  podcidr=$(kubeadm config view | grep podSubnet | awk -F": " '{print$2}')
  rm -rf calico.yaml
  cp yamls/calico.yaml calico.yaml
  sed -i "s:POD_CIDR:${podcidr}:g" calico.yaml
  kubectl create -f calico.yaml
  rm -rf calico.yaml
}

setupCalico
