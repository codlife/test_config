FROM openjdk:7
MAINTAINER wuyuewen@otcaix.iscas.ac.cn

RUN apt-get update && DEBIAN_FRONTEND=noninteractive apt-get -y install openssh-server net-tools vim bc
RUN mkdir -p /var/run/sshd
RUN mkdir -p /root/.ssh
RUN mkdir -p /var/run/sshd && sed -i "s/UsePrivilegeSeparation.*/UsePrivilegeSeparation no/g" /etc/ssh/sshd_config && sed -i "s/UsePAM.*/UsePAM no/g" /etc/ssh/sshd_config && sed -i "s/PermitRootLogin.*/PermitRootLogin yes/g" /etc/ssh/sshd_config

ADD set_root_pw.sh /set_root_pw.sh
ADD run.sh /run.sh
RUN chmod +x /*.sh

ENV AUTHORIZED_KEYS **None**

EXPOSE 22
CMD ["/run.sh"]
