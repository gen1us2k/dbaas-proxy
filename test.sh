#!/bin/bash
export KUBECONFIG_DATA=$(cat ~/.kube/config | base64 )
data="{\"name\": \"minikube\", \"kubeconfig\": \"$KUBECONFIG_DATA\"}"
echo $data

curl http://localhost:8080/k8s \
	-H "Content-type: application/json"\
	-XPOST \
	--data "$data"

