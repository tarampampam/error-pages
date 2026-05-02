# Local Kubernetes on Linux

A short guide to getting `kind` running on your Linux box, spinning up a cluster, doing useful stuff, and nuking 
everything back to zero when you're done.

`kind` = **K**ubernetes **IN** **D**ocker. Each "node" is just a Docker container running a real kubelet. No VMs,
no hypervisor - just Docker.

## Prereqs

- Docker installed, your user in the `docker` group ([how to install][install-docker] + 
  [post-install steps][postinstall-docker])
- `kubectl` - `kind` itself doesn't need it, but you definitely do ([how to install][install-kubectl])
- `kind` itself, `v0.31.0` is used in this guide ([how to install][install-kind])
- `helm` - we will use it to deploy some apps later ([how to install][install-helm])

[install-docker]:https://docs.docker.com/engine/install/debian/#install-using-the-convenience-script
[postinstall-docker]:https://docs.docker.com/engine/install/linux-postinstall/#manage-docker-as-a-non-root-user
[install-kubectl]:https://kubernetes.io/docs/tasks/tools/install-kubectl-linux/
[install-kind]:https://kind.sigs.k8s.io/docs/user/quick-start/#installing-from-release-binaries
[install-helm]:https://helm.sh/docs/intro/install/

## Spin up a cluster

The `kind create cluster` command creates a cluster with no host port mappings. Traefik, Nginx, and other ingress
controllers that expect to bind ports 80 and 443 on the host won't work with the default kind config, which is why
the cluster must be created with `extraPortMappings` instead.

Save the following to `cluster-config.yaml`:

```yaml
# File: cluster-config.yaml
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  kubeadmConfigPatches:
  - |
    kind: InitConfiguration
    nodeRegistration:
      kubeletExtraArgs:
        node-labels: "ingress-ready=true"
  extraPortMappings:
  - {containerPort: 80,  hostPort: 80,  protocol: TCP}
  - {containerPort: 443, hostPort: 443, protocol: TCP}
```

- `extraPortMappings` forwards host ports 80 and 443 into the kind node container so that the cluster's ingress is
  reachable directly from your machine.
- `ingress-ready=true` is the node label that the ingress `nodeSelector` targets.

```shell
$ kind create cluster --name error-pages-test-cluster --config cluster-config.yaml
Creating cluster "error-pages-test-cluster" ...
 ✓ Ensuring node image (kindest/node:v1.35.0) 🖼 
 ✓ Preparing nodes 📦  
 ✓ Writing configuration 📜 
 ✓ Starting control-plane 🕹️ 
 ✓ Installing CNI 🔌 
 ✓ Installing StorageClass 💾 
Set kubectl context to "kind-error-pages-test-cluster"
You can now use your cluster with:

kubectl cluster-info --context kind-error-pages-test-cluster

Thanks for using kind! 😊
```

That's it. `kind` boots a Docker container that acts as a k8s node, brings up the control plane, and writes a
kubeconfig context to `~/.kube/config`. It takes ~30 seconds.

```shell
$ cat ~/.kube/config | grep -A5 'contexts:'
contexts:
- context:
    cluster: kind-error-pages-test-cluster
    user: kind-error-pages-test-cluster
  name: kind-error-pages-test-cluster
current-context: kind-error-pages-test-cluster
```

## Verify it actually works

```shell
$ kubectl cluster-info --context kind-error-pages-test-cluster
Kubernetes control plane is running at https://127.0.0.1:41711
CoreDNS is running at https://127.0.0.1:41711/api/v1/namespaces/kube-system/services/kube-dns:dns/proxy

$ kubectl get nodes
NAME                                    STATUS   ROLES           AGE     VERSION
error-pages-test-cluster-control-plane  Ready    control-plane   3m52s   v1.35.0

$ kubectl get pods -A
NAMESPACE            NAME                                                             READY   STATUS    RESTARTS   AGE
kube-system          coredns-7d764666f9-jq529                                         1/1     Running   0          75s
kube-system          coredns-7d764666f9-wrmt6                                         1/1     Running   0          75s
kube-system          etcd-error-pages-test-cluster-control-plane                      1/1     Running   0          82s
kube-system          kindnet-zjqch                                                    1/1     Running   0          75s
kube-system          kube-apiserver-error-pages-test-cluster-control-plane            1/1     Running   0          82s
kube-system          kube-controller-manager-error-pages-test-cluster-control-plane   1/1     Running   0          81s
kube-system          kube-proxy-75jlw                                                 1/1     Running   0          75s
kube-system          kube-scheduler-error-pages-test-cluster-control-plane            1/1     Running   0          82s
local-path-storage   local-path-provisioner-67b8995b4b-4whf7                          1/1     Running   0          75s
```

### Shell into a node (debugging)

```shell
$ docker exec -it error-pages-test-cluster-control-plane bash
crictl ps # actual containers running on that "node"
crictl images
```

## Cleanup

### Delete the cluster

```shell
$ kind get clusters
error-pages-test-cluster

$ kind delete cluster --name error-pages-test-cluster
Deleting cluster "error-pages-test-cluster" ...
Deleted nodes: ["error-pages-test-cluster-control-plane"]
```

## Common gotchas

- **Pods can't pull images** → forgot `kind load docker-image`, or using `:latest`.
- **`kubectl` talks to the wrong cluster** → check `kubectl config current-context`. It might still point at a 
  deleted kind context.
- **Cluster creation hangs** → usually Docker resource limits or low RAM. Check `dmesg` for OOM kills.
- **API server port conflicts** → by default kind picks a random host port for the API. Pin it via
  `networking.apiServerPort` in the config file.
- **DNS doesn't resolve inside pods** → restart CoreDNS: `kubectl -n kube-system rollout restart deploy/coredns`.
