
## 1. Steps

- vi images.conf (to modify the all images version if you need)
- bash pull.sh （to pull down all images ）
  - for master: bash dev.sh (to install kubernetes on develop enviroment) or bash prod.sh (to kubernetes on production enviroment, we will support this script later)
  - for node: bash join.sh (to join to the master)
- select one of network plugins to continue install kubernetes. see [README.md](../kube-networks/README.md)

## 2. Reference
- https://github.com/mritd/gcr
