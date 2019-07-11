#! /bin/bash
#上传代码到hdfs
echo "hadoop fs -copyFromLocal $1 $2/code"
hadoop fs -copyFromLocal $1 $2/code
#执行脚本


