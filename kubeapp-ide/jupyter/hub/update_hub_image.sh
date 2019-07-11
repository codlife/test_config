echo "build hub image"
docker build . -t jupyterhub/k8s-hub:0.8.0
cd ~
echo "save image"
docker save jupyterhub/k8s-hub:0.8.0 > jupyterhub_k8s-hub_0.8.0.tar
echo "transport image to node"
scp ./jupyterhub_k8s-hub_0.8.0.tar root@slave1:~/images/
scp ./jupyterhub_k8s-hub_0.8.0.tar root@slave2:~/images/

