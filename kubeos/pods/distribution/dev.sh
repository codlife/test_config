###########################################
##
## Copyright (2019,) Institute of Software
##       Chinese Academy of Sciences
##        wuheng@otcaix.iscas.ac.cn  
##
##########################################

kubectl create -f yamls/registry-insec.yaml
kubectl create -f yamls/registry-svc.yaml

openssl req -subj "/C=CN/ST=Beijing/L=Beijing/O=Cloudplus/OU=Cloudplus/CN=registry.com/emailAddress=registry@registry.com"  -newkey rsa:4096 -nodes -sha256 -keyout domain.key   -x509 -days 365 -out domain.crt

#mkdir -p /etc/kubernetes/registry/{images,certs,auth}
#docker run --rm --entrypoint htpasswd registry:2.7.1 -Bbn rob 1234 > /etc/kubernetes/registry/auth/htpasswd

# docker run -d --rm -p 8080:8080 -e REG1=http://192.168.44.129:5000/v2/  atcol/docker-registry-ui
# yum install letsencrypt
# letsencrypt certonly -d registry.com
# /etc/letsencrypt/live/registry.com

#rm -rf certs
#mkdir certs
#cd certs
#openssl genrsa -out server.key 2048
#openssl req -subj "/C=CN/ST=Beijing/L=Beijing/O=Cloudplus/OU=Cloudplus/CN=registry.com/emailAddress=registry@registry.com" -new -key server.key -out server.csr
#openssl x509 -req -days 3650 -in server.csr -signkey server.key -out server.crt
#openssl genrsa -out ca.key 2048
#openssl req -subj "/C=CN/ST=Beijing/L=Beijing/O=Cloudplus/OU=Cloudplus/CN=registry.com/emailAddress=registry@registry.com" -new -x509 -days 3650 -key ca.key -out ca.crt
#openssl genrsa -out server.key 2048
#openssl req -subj "/C=CN/ST=Beijing/L=Beijing/O=Cloudplus/OU=Cloudplus/CN=registry.com/emailAddress=registry@registry.com" -new -key server.key -out server.csr
#openssl x509 -req -days 3650 -in server.csr -CA ca.crt -CAkey ca.key -set_serial 01 -out server.crt

#\cp * /etc/kubernetes/registry/certs
