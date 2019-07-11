#!/bin/bash
rm -f /root/scripts/registerClient
rm -f /root/scripts/start-worker-hadoop.sh
/usr/sbin/sshd -D &
#ip=`ifconfig eth0 | grep 'inet addr' | cut -d : -f 2 | cut -d ' ' -f 1`
ip=`ifconfig eth0 | grep 'inet' | awk '{print $2}'`
sed -i "s/hadoop-master/$ip/" $HADOOP_HOME/etc/hadoop/core-site.xml
sed -i "s/hadoop-master/$ip/" $HADOOP_HOME/etc/hadoop/yarn-site.xml

$HADOOP_HOME/sbin/start-dfs.sh &

$HADOOP_HOME/sbin/start-yarn.sh &

$SPARK_HOME/bin/spark-class org.apache.spark.deploy.master.Master --port 7077 --webui-port 8080 &

/root/scripts/registerServer &

/bin/upload/upload &

/bin/gotty --port 8000 --permit-write --reconnect /bin/bash
