## 1. 研发背景

将可执行文件打包成rpm，使得该文件可以容易的在CentOS/SUSE节点之间交换。

## 2. 设计原则

1. 通用性：使用标准的rpm，方便在CentOS/SUSE机器上安装，且ubuntu通过插件也能支持
2. 简洁化：容易使用，不能增加学习复杂度
3. 一致性：尽量做到与Docker的镜像管理一致

## 3 执行步骤

### 3.1基于源代码的方法

1. 将等待打包的文件拷贝到本目录（本目录下有rpmbuilder, build.spec等文件），注意文件名为Name-Version，必须与RpmFile中NAME和VERSION一致

2. 编辑RpmFile

```
#名字，必填
NAME apache-tomcat
#版本号，必填
VERSION 8.5.35
#GIT号，选填，不填写则为1811
GIT 1811
#选填，默认为wuheng@iscas.ac.cn
MAINTAINER wuheng@iscas.ac.cn
#选填，可以多个，注意路径的完整性
CMD echo "Hello"
#选填写，可以多个
#ENV A=B C=D
#必填，至少一个，类似与Dockerfile的From
DEP java-1.8.0-openjdk net-tools
#选填
#EXPOSE 8080
#选填，默认为/opt，是指该rpm安装到目标机器的根目录
WORKDIR /opt/abc
#必填 相对路径
START bash ./bin/startup.sh
STOP  bash ./bin/shutdown.sh
```

3. 生成RPM的SPEC文件

```
bash rpmbuilder parse RpmFile
```

4. 编译成rpm

```
bash rpmbuilder build RpmFile
```

5.从/root/rpmbuild/RPMS/x86_64/获取打好包的文件

6.转deb包

```
alien *.rpm
```
