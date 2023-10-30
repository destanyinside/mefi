# mefi
- `Mefi` - Multi-cluster extension for ingress, based on endpoints, which are replicated cluster by cluster
- Without any mash and other things overhead

## Required
- Kubernetes clusters without overlay networks. Need RIP or BGP announced pod cidr in network
- `Mefi` must be running in all clusters that need balancing ingress traffic
- Endpoints that need to balance ingress traffic must have a label: `default is isMefiRemote=true, param --remote-filter`
- Endpoints for same application must have same name in all clusters 

## How it's work
1. Create RBAC for [local](https://github.com/destanyinside/mefi/tree/main/deploy/chart/templates/local_rbac.yaml) and [remote](https://github.com/destanyinside/mefi/tree/main/deploy/chart/templates/remote_rbac.yaml) in all clusters
2. Prepare config for `mefi`, example [config](hack/config.yaml)
3. Run `mefi` in all clusters that need balancing ingress traffic. [Helm chart](https://github.com/destanyinside/mefi/tree/main/deploy/chart)
4. Create kubernetes service with label `isMefiRemote=true (default, param --remote-filter)` for your pods
5. `Mefi` watch and replicate all endpoints from remote clusters in local namespace `mefi-system (default, param --mefi-namespace)` with name `<remote_endpoint_name>-<cluster_name>`
   + 5.1. Added label `(default isMefiOriginalName, param --original-filter)` with original name
   + 5.2. Added label `(default isMefiLocal=true, param --local-filter)`
6. `Mefi` watch replicated endpoints in local cluster with label `(default isMefiLocal=true, param --local-filter)`
   + 6.1. `Mefi` receive event with create/update/delete endpoints
   + 6.2. `Mefi` get all endpoints with label `(default isMefiOriginalName, param --original-filter)`
   + 6.3. Create/update/delete endpoints in local cluster with merged subsets from replicated endpoints and name based from label value `(default isMefiOriginalName, param --original-filter)`
7. Prepare your ingress resource (ingress, httpProxy, ingressRoute, etc) and service (preferably without key spec.selector) in namespace `mefi-system (default, param --mefi-namespace)`
8. Profit! You will take multi-cluster balancing ingress traffic
