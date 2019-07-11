###########################################
##
## Copyright (2019,) Institute of Software
##       Chinese Academy of Sciences
##        wuheng@otcaix.iscas.ac.cn  
##
##########################################

VALUE=gcr.azk8s.cn
KEY=gcr.io

while read line
do
    img=$(echo $line | awk -F":" '{print$1}')
    ver=$(echo $line | awk -F":" '{print$2}')
    res=$(docker images | grep "$img" | grep "$ver" | grep -v grep)
    if [[ -z $res ]]
    then
      name=${line//$KEY/$VALUE}
      echo docker pull $name
      docker pull $name
      docker tag $name $line
      docker rmi $name
    fi
done < images.conf
