
setupHelm()
{
  cp ../helm/helm /usr/bin/
  kubectl --namespace kube-system create serviceaccount tiller
  kubectl create clusterrolebinding tiller --clusterrole cluster-admin --serviceaccount=kube-system:tiller
  helm init --service-account tiller --wait
  kubectl patch deployment tiller-deploy --namespace=kube-system --type=json --patch='[{"op": "add", "path": "/spec/template/spec/containers/0/command", "value": ["/tiller", "--listen=localhost:44134"]}]'
}

setUpJupyter()
{
  helm repo add jupyterhub https://jupyterhub.github.io/helm-chart/
  helm repo update

  hex=$(openssl rand -hex 32)
  rm -rf yamls/config.yaml
  echo "proxy:" >> yamls/config.yaml
  echo "  secretToken: \"$hex\"" >> yamls/config.yaml

  RELEASE=jhub
  NAMESPACE=jhub

  helm upgrade --install $RELEASE jupyterhub/jupyterhub --namespace $NAMESPACE  --version=0.8.0 --values yamls/config.yaml --timeout=3000
}

setupHelm
setUpJupyter
