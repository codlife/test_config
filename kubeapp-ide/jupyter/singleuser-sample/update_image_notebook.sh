echo "build image"
docker build . -t jupyterhub/k8s-singleuser-sample:0.8.0
echo "save image"
cd ~
docker save jupyterhub/k8s-singleuser-sample:0.8.0 > jupyterhub_k8s-singleuser-sample_0.8.0.tar
echo "tranport image to nodes"
scp ./jupyterhub_k8s-singleuser-sample_0.8.0.tar root@slave1:~/images/
scp ./jupyterhub_k8s-singleuser-sample_0.8.0.tar root@slave2:~/images/

