############################################
##
## Copyright (2019, ) Institute of Software
##      Chinese Academy of Sciences
##       wuheng@otcaix.iscas.ac.cn
##
############################################

function setupDashboard()
{
  kubectl create -f yamls/kubedash.yaml
  kubectl create -f yamls/kubedash-admin.yaml
  kubectl create -f yamls/kubedash-role.yaml
  port=$(kubectl -n kube-system get service kubernetes-dashboard | awk 'NR>1 {print$5}'| awk -F"/" '{print$1}' | awk -F":" '{print$2}')
  echo "URL: https://<masterip>:"$port  
  echo $(kubectl -n kube-system describe secret $(kubectl -n kube-system get secret | grep admin-user | awk '{print $1}') | grep "token:") 
}

setupDashboard
