#!/usr/bin/python
# -*- coding: utf-8 -*-
"""
@author:
@contact:
@time:
"""
from __future__ import print_function
from pyspark.sql import SparkSession
import os, time

if __name__ == "__main__":
   # 设置spark_home环境变量，路径不能有中文、空格
   #os.environ['SPARK_HOME'] = "E:/data_page/spark-2.0.2-bin-hadoop2.7"
   # 运行在本地（local），2个线程，一行写不完换行时用“\”
   spark = SparkSession.builder\
      .master("spark://hmaster:7077")\
      .appName("test")\
      .getOrCreate()
   # 如果想看函数源码，可以通过ctrl+点击函数的形式跳转到函数详情界面
   datas = ["hi I love you", "hello", "ni hao"]
   sc = spark.sparkContext
   rdd = sc.parallelize(datas)
   # 查看数据类型 type()
   print(type(datas))
   print(type(rdd))
   #获取总数，第一条数据
   print(rdd.count())
   print(rdd.first())
   # 每个spark运行会有一个监控界面（WEB UI4040），为了监控，让线程休眠一段时间，然后打开localhost:4040页面
   spark.stop()

