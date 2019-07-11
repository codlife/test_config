############################################
##
## Copyright (2019, ) Institute of Software
##      Chinese Academy of Sciences
##       wuheng@otcaix.iscas.ac.cn
##
############################################

function setupVpp()
{
  #ip=$(cat /etc/kubernetes/admin.conf  | grep server | awk -F"//" '{print$2}' | awk -F":" '{print$1}')
  #nic=$(ip a | grep "$ip" | awk '{print$NF}')
  #bus=$(lshw -class network -businfo | grep "$nic" | awk '{print$1}' | awk -F"@" '{print$2}')
  #echo $bus
  systemctl start vpp
  systemctl enable vpp
  kubectl apply -f yamls/contiv-vpp.yaml
  kubectl apply -f yamls/contiv-vpp-ui.yaml
}

setupVpp
