FROM centos:centos7
MAINTAINER wuyuewen@otcaix.iscas.ac.cn
WORKDIR /root

# install packages
RUN yum install -y openssh-server net-tools vim bc iproute sed git unzip
RUN yum clean all
RUN rm -fr /tmp/*

# install JDK 8
COPY jdk-8u144-linux-x64.rpm /root
RUN rpm -ivh jdk-8u144-linux-x64.rpm && \
    rm jdk-8u144-linux-x64.rpm

# install Scala
COPY scala-2.11.12.rpm /root
RUN rpm -ivh scala-2.11.12.rpm && \
    rm scala-2.11.12.rpm

# install hadoop 2.7.6
COPY hadoop-2.7.6.tar.gz /root 
RUN tar -xzvf hadoop-2.7.6.tar.gz && \
    mv hadoop-2.7.6 /usr/local/hadoop && \
    rm hadoop-2.7.6.tar.gz

# install spark 2.2.0
COPY spark-2.2.0.zip /root
RUN unzip spark-2.2.0.zip && \
    mv spark-2.2.0 /usr/local/spark && \
    rm spark-2.2.0.zip

# set environment variable
ENV JAVA_HOME=/usr/java/jdk1.8.0_144
ENV HADOOP_HOME=/usr/local/hadoop 
ENV PATH=$PATH:/usr/local/hadoop/bin:/usr/local/hadoop/sbin
ENV SPARK_HOME=/usr/local/spark
ENV PATH=$PATH:/usr/local/spark/bin

# ssh without key
RUN mkdir -p /var/run/sshd && \
    sed -i "s/#UseDNS.*/UseDNS no/g" /etc/ssh/sshd_config && \
    sed -i "s/UsePAM.*/UsePAM no/g" /etc/ssh/sshd_config && \
    sed -i "s/#UsePrivilegeSeparation.*/UsePrivilegeSeparation no/g" /etc/ssh/sshd_config && \
    sed -i "s/#PermitRootLogin.*/PermitRootLogin yes/g" /etc/ssh/sshd_config && \
    echo 'root:123456' | chpasswd && \
    ssh-keygen -t rsa -f /etc/ssh/ssh_host_rsa_key -P "" && \
    ssh-keygen -t ecdsa -f /etc/ssh/ssh_host_ecdsa_key -P "" && \
    ssh-keygen -t ed25519 -f /etc/ssh/ssh_host_ed25519_key -P "" && \
    ssh-keygen -t rsa -f ~/.ssh/id_rsa -P "" && \
    cat ~/.ssh/id_rsa.pub >> ~/.ssh/authorized_keys
    
# create dirs
RUN mkdir -p ~/hdfs/namenode && \ 
    mkdir -p ~/hdfs/datanode && \
    mkdir $HADOOP_HOME/logs  && \
    mkdir /bin/upload &&\
    mkdir /root/scripts

# copy necessary scripts
COPY config/* /root/scripts/
COPY gotty /bin/
COPY upload/* /bin/upload/

# configuration files in Hadoop
RUN mv /root/scripts/hadoop-env.sh /usr/local/hadoop/etc/hadoop/hadoop-env.sh && \
    mv /root/scripts/ssh_config ~/.ssh/config && \
    mv /root/scripts/hdfs-site.xml $HADOOP_HOME/etc/hadoop/hdfs-site.xml && \ 
    mv /root/scripts/core-site.xml $HADOOP_HOME/etc/hadoop/core-site.xml && \
    mv /root/scripts/mapred-site.xml $HADOOP_HOME/etc/hadoop/mapred-site.xml && \
    mv /root/scripts/yarn-site.xml $HADOOP_HOME/etc/hadoop/yarn-site.xml && \
    mv /root/scripts/slaves $HADOOP_HOME/etc/hadoop/slaves && \
    mv /root/scripts/spark-env.sh $SPARK_HOME/conf && \
    mv /root/scripts/run.sh ~/run.sh && \
    mv /root/scripts/run-wordcount.sh ~/run-wordcount.sh

# modify files' permission in Hadoop
RUN chmod +x /root/scripts/start-master-hadoop.sh && \
    chmod +x /root/scripts/start-worker-hadoop.sh && \
    chmod +x /root/scripts/registerServer && \
    chmod +x /root/scripts/registerClient && \
    chmod 775 ~/run.sh && \
    chmod +x ~/run-wordcount.sh && \
    chmod +x $HADOOP_HOME/sbin/start-dfs.sh && \
    chmod +x $HADOOP_HOME/sbin/start-yarn.sh && \
    chmod +x /bin/gotty

# modify directory's permission in Spark
#RUN chmod -R 777 $SPARK_HOME

# format namenode in Hadoop
RUN /usr/local/hadoop/bin/hdfs namenode -format

# install HiBench
#COPY HiBench.zip /root
#RUN unzip HiBench.zip && \
#    rm HiBench.zip

EXPOSE 22
CMD ["/usr/sbin/sshd", "-D"]
#ENTRYPOINT ["/usr/sbin/init"]
#CMD ["/usr/sbin/init"]
