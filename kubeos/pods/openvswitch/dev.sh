#############################################
##
## Copyright (2019, ) Institute of Software
##     Chinese Academy of Sciences
##      wuheng@otcaix.iscas.ac.cn
##
############################################

systemctl start openvswitch
systemctl enable openvswitch

\cp bridge /opt/cni/bin/bridge

active=$(systemctl status openvswitch | grep Active | awk '{print$2}')

if [[ $active != "active" ]]
then
  echo "please install openvswitch"
  exit
fi

kubectl apply -f ../flannel/yamls/kube-flannel.yml
kubectl apply -f yamls/multus.yaml 
kubectl apply -f yamls/ovs-cni.yaml 
kubectl apply -f yamls/ovs-conf.yaml
