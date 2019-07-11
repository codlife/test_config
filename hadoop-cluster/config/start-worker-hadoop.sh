#!/bin/bash
rm -f /root/scripts/start-master-hadoop.sh
rm -f /root/scripts/registerServer
/usr/sbin/sshd -D &
sed -i "s/hadoop-master/$1/" $HADOOP_HOME/etc/hadoop/core-site.xml
sed -i "s/hadoop-master/$1/" $HADOOP_HOME/etc/hadoop/yarn-site.xml

$HADOOP_HOME/sbin/hadoop-daemon.sh start datanode & 
$HADOOP_HOME/sbin/yarn-daemon.sh start nodemanager &

$SPARK_HOME/bin/spark-class org.apache.spark.deploy.worker.Worker spark://$1:7077 --webui-port 8081 &

/root/scripts/registerClient $1

tail -f /dev/null
