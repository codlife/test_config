#!/bin/sh
NAME=$1
cat > /home/henry/kubeapp-ide/jupyter/yamls/user-$NAME-pv.yaml << EOF
kind: PersistentVolume
apiVersion: v1
metadata:
  name: $NAME-pv
  namespace: jhub
  labels:
    appname: jupyterhub
spec:
  capacity:
    storage: 20Gi
  accessModes:
    - ReadWriteOnce
  hostPath:
    path: "/opt/code/BDP/data/private/$NAME"
EOF
kubectl create -f /home/henry/kubeapp-ide/jupyter/yamls/user-$NAME-pv.yaml
mkdir /opt/code/BDP/data/private/$NAME
chmod 777 /opt/code/BDP/data/private/$NAME


