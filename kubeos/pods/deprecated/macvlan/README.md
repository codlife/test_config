## 0. Note

- this project can only work in a container (CentOS 7)
- if you only has a wireless NIC, this project cannot work well
- All machine with the same type NIC

## 1. Steps

-  bash build.sh  
-  bash dev.sh (to install kubernetes on develop enviroment) or bash prod.sh (to kubernetes on production enviroment, we will support this script later)

Note that current work can only support an available NIC, if you have mutiple NICs, please config yamls/macvlan.yaml.
Here is an example, at least modify the MACVLAN_NIC's value

```
env:
   - name: MACVLAN_NIC
     value: wlp3s0
   - name: MACVLAN_GATEWAY
     value: 192.168.1.1
   - name: MACVLAN_SUBNET
     value: 192.168.1.5/16
   - name: MACVLAN_RANGESTART
     value: 192.168.0.0
   - name: MACVLAN_RANGEEND
     value: 192.168.255.255
```

## 2. Reference
- https://github.com/containernetworking/plugins/releases
- https://blog.csdn.net/cloudvtech/article/details/79830887
