FROM centos:latest

MAINTAINER wuheng wuheng@iscas.ac.cn

RUN  echo "67.219.148.138 mirrorlist.centos.org" >> /etc/hosts
#RUN  yum install rpm-build -y
#RUN  yum install gcc gcc-c++ -y
#RUN  yum install glibc -y

ADD  build.spec /root/build.spec
CMD  mkdir -p /root/rpmbuild
