#!/bin/bash

kubectl create -f namespace-hadoop.yaml
kubectl create -f hadoop-master-controller.yaml 
kubectl create -f hadoop-master-service.yaml
kubectl create -f hadoop-worker-controller.yaml
