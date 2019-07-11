############################################
##
## Copyright (2019, ) Institute of Software
##      Chinese Academy of Sciences
##       wuheng@otcaix.iscas.ac.cn
##
############################################

kubectl apply -f yamls/crd-10.yaml 
kubectl apply -f yamls/crd-11.yaml 
kubectl apply -f yamls/crd-certmanager-10.yaml 
kubectl apply -f yamls/crd-certmanager-11.yaml 
kubectl apply -f yamls/namespace.yaml 
kubectl apply -f yamls/istio.yaml
