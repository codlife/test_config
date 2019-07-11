echo "build flask app image"
docker build . -t k8s_flask_app_example:0.1
cd ~
echo "save image"
docker save k8s_flask_app_example:0.1 > k8s_flask_app_example_0.1.tar
echo "transport image to node"
scp ./k8s_flask_app_example_0.1.tar root@slave1:~/images/
scp ./k8s_flask_app_example_0.1.tar root@slave2:~/images/
ssh root@slave1
docker load < ~/images/k8s_flask_app_example_0.1.tar
exit
ssh root@slave2
docker load < ~/images/k8s_flask_app_example_0.1.tar
exit
