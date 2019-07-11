############################################
##
## Copyright (2019, ) Institute of Software
##      Chinese Academy of Sciences
##       wuheng@otcaix.iscas.ac.cn
##
############################################

function setupDashboard()
{
  ./grctl init --iip <必须指定内网ip> --eip <可选外网ip> --deploy-type thirdparty
}

## ca: cat /etc/kubernetes/pki/ca.crt
## token: kubectl -n kube-system get secret | grep kube-proxy | awk '{print "secret/"$1}' | xargs kubectl describe -n kube-system | grep token: | awk -F: '{print $2}' | xargs echo

setupDashboard
