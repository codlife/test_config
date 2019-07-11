############################################
##
## Copyright (2019, ) Institute of Software
##      Chinese Academy of Sciences
##       wuheng@otcaix.iscas.ac.cn
##
############################################

function setupPrometheus()
{
  kubectl apply -f yamls/kube-fluentd.yaml
}

setupPrometheus
