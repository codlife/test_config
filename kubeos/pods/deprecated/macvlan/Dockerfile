FROM centos:7

MAINTAINER wuheng@otcaix.iscas.ac.cn

COPY ./node/10-macvlan.conf.template \
     ./node/macvlan \
     ./node/host-local \
     ./node/loopback \
     ./node/cni.conf \
     ./node/macvlan.py \
     /home/node/

WORKDIR /home/node

CMD ["/usr/bin/python", "macvlan.py"]
