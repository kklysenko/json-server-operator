# json-server-operator

To run the project locally:
Prerequisites:
- [kind] (https://kind.sigs.k8s.io/)
- [kubebuilder] (https://book.kubebuilder.io/)

1. Make sure your kubectl context is directed to your kind cluster.
2. Apply CRD `make install`
3. Run `make run`
4. Apply JsonServer example `kubectl apply -f config/samples/v1_jsonserver.yaml`

To deploy the project to Kind cluster:
Prerequisites:
- [kind] (https://kind.sigs.k8s.io/)
- [kubebuilder] (https://book.kubebuilder.io/)
- [cert-manager] (https://cert-manager.io/docs/installation/)

1. Make sure your kubectl context is directed to your kind cluster.
2. Run `make deploy-kind`
3. Apply JsonServer example `kubectl apply -f config/samples/v1_jsonserver.yaml`

To deploy the project to any cluster:
Prerequisites:
- [kubebuilder] (https://book.kubebuilder.io/)
- [cert-manager] (https://cert-manager.io/docs/installation/)
- k8s cluster

1. Make sure your kubectl context is directed to your kind cluster.
2. Run `make deploy-ttl IMG=ttl.sh/your_unique_image_id:0.0.1`
3. Apply JsonServer example `kubectl apply -f config/samples/v1_jsonserver.yaml`
