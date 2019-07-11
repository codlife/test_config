#! /bin/sh
for host in 'root@slave001' 'root@slave002' 'root@slave003' 'root@slave004'
do
   for image in 'jupyter/base-notebook:7f1482f5a136' 'jupyterhub/k8s-network-tools:0.8.0' 'jupyterhub/configurable-http-proxy:3.0.0' 
   do
      ssh -t -p 22 ${host} "docker pull ${image}"
   
   done
   
done
