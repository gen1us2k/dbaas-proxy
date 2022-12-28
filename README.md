# dbaas-proxy
It's a PoC to move kubectl calls to frontend 

## API Design

API Design is simple

```
POST /k8s - Create kubernetes cluster. It'll add to the in-memory storage
DELETE /k8s/:name - delete kubernetes cluster
ANY /proxy/:name - Proxy endpoint to a registered k8s cluster 
```

## Testing

Spin up a kubernetes cluster with minikube or using managed services (EKS, GCP) and register it against API

```bash
#!/bin/bash
export KUBECONFIG_DATA=$(cat ~/.kube/config | base64 )
data="{\"name\": \"minikube\", \"kubeconfig\": \"$KUBECONFIG_DATA\"}"
echo $data

curl http://localhost:8080/k8s \
	-H "Content-type: application/json"\
	-XPOST \
	--data "$data"

```
Create a sample kubeconfig in ~/.kube/proxy

```
apiVersion: v1
clusters:
- cluster:
    server: http://localhost:8080/proxy/minikube
  name: minikube
contexts:
- context:
    cluster: minikube
    user: ""
  name: default
- context:
    cluster: minikube
    extensions:
    - extension:
        last-update: Wed, 28 Dec 2022 15:39:44 +06
        provider: minikube.sigs.k8s.io
        version: v1.28.0
      name: context_info
    namespace: default
    user: minikube
  name: minikube
current-context: minikube
kind: Config
preferences: {}
```

Run kubectl

```
kubectl --kubeconfig ~/.kube/proxy get po -A
NAMESPACE     NAME                                  READY   STATUS    RESTARTS   AGE
kube-system   coredns-64897985d-2r5pn               1/1     Running   0          39m
kube-system   etcd-minikube                         1/1     Running   0          40m
kube-system   kindnet-5r6zk                         1/1     Running   0          39m
kube-system   kindnet-gt7vm                         1/1     Running   0          39m
kube-system   kindnet-jql2p                         1/1     Running   0          38m
kube-system   kindnet-ksb2s                         1/1     Running   0          39m
kube-system   kindnet-rdbc4                         1/1     Running   0          39m
kube-system   kube-apiserver-minikube               1/1     Running   0          40m
kube-system   kube-controller-manager-minikube      1/1     Running   0          40m
kube-system   kube-proxy-6sck6                      1/1     Running   0          38m
kube-system   kube-proxy-7dsj9                      1/1     Running   0          39m
kube-system   kube-proxy-mfgt6                      1/1     Running   0          39m
kube-system   kube-proxy-sjmj7                      1/1     Running   0          39m
kube-system   kube-proxy-wg2cb                      1/1     Running   0          39m
kube-system   kube-scheduler-minikube               1/1     Running   0          40m
kube-system   kubevirt-hostpath-provisioner-d8nrv   1/1     Running   0          38m
kube-system   kubevirt-hostpath-provisioner-flrmb   1/1     Running   0          38m
kube-system   kubevirt-hostpath-provisioner-nsb7p   1/1     Running   0          38m
kube-system   kubevirt-hostpath-provisioner-vwpt6   1/1     Running   0          38m
kube-system   kubevirt-hostpath-provisioner-xdcl8   1/1     Running   0          38m
```
