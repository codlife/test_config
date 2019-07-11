###########################################
##
## Copyright (2019,) Institute of Software
##       Chinese Academy of Sciences
##        wuheng@otcaix.iscas.ac.cn  
##
##########################################

VALUE1=gcr.azk8s.cn
KEY1=gcr.io

VALUE2=gcr.azk8s.cn/google_containers
KEY2=k8s.gcr.io

function download()
{
  name=${line//$1/$2}
  echo docker pull $name
  docker pull $name
  docker tag $name $3
  docker rmi $name
}

while read line
do
    img=$(echo $line | awk -F":" '{print$1}')
    ver=$(echo $line | awk -F":" '{print$2}')
    res=$(docker images | grep "$img" | grep "$ver" | grep -v grep)
    if [[ -z $res ]]
    then
      prefix=$(echo $line | awk -F"/" '{print$1}')
      if [[ "$prefix" == "$KEY1" ]]
      then
        download $KEY1 $VALUE1 $line
      elif [[ "$prefix" == "$KEY2" ]]
      then
        download $KEY2 $VALUE2 $line
      fi
    fi
done < images.conf
