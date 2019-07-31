## Available kubectl plugins

The following kubectl plugins are currently available
on [Krew plugin index](https://sigs.k8s.io/krew-index).

To install these kubectl plugins:

1. [Install Krew](https://github.com/kubernetes-sigs/krew#installation)
2. Run `kubectl krew install PLUGIN_NAME` to install a plugin via Krew.

Name | Description | Stars
---- | ----------- | -----
[access-matrix](https://github.com/corneliusweig/rakkess) | Show an RBAC access matrix for server resources | ![GitHub stars](https://img.shields.io/github/stars/corneliusweig/rakkess.svg?label=stars&logo=github)
[bulk-action](https://github.com/emreodabas/kubectl-plugins#kubectl-bulk) | Do bulk actions on Kubernetes resources. | ![GitHub stars](https://img.shields.io/github/stars/emreodabas/kubectl-plugins.svg?label=stars&logo=github)
[ca-cert](https://github.com/ahmetb/kubectl-extras) | Print the PEM CA certificate of the current cluster | ![GitHub stars](https://img.shields.io/github/stars/ahmetb/kubectl-extras.svg?label=stars&logo=github)
[change-ns](https://github.com/juanvallejo/kubectl-ns) | View or change the current namespace via kubectl. | ![GitHub stars](https://img.shields.io/github/stars/juanvallejo/kubectl-ns.svg?label=stars&logo=github)
[config-cleanup](https://github.com/b23llc/kubectl-config-cleanup) | Automatically clean up your kubeconfig | ![GitHub stars](https://img.shields.io/github/stars/b23llc/kubectl-config-cleanup.svg?label=stars&logo=github)
[cssh](https://github.com/containership/kubectl-cssh) | SSH into Kubernetes nodes | ![GitHub stars](https://img.shields.io/github/stars/containership/kubectl-cssh.svg?label=stars&logo=github)
[debug-shell](https://github.com/danisla/kubefunc) | Create pod with interactive kube-shell. | ![GitHub stars](https://img.shields.io/github/stars/danisla/kubefunc.svg?label=stars&logo=github)
[exec-as](https://github.com/jordanwilson230/kubectl-plugins/tree/krew#kubectl-exec-as) | Like kubectl exec, but offers a `user` flag to exec as root or any other user. | ![GitHub stars](https://img.shields.io/github/stars/jordanwilson230/kubectl-plugins.svg?label=stars&logo=github)
[exec-cronjob](https://github.com/thecloudnatives/kubectl-plugins#exec-cronjob) | Run a CronJob immediately as Job | ![GitHub stars](https://img.shields.io/github/stars/thecloudnatives/kubectl-plugins.svg?label=stars&logo=github)
[get-all](https://github.com/corneliusweig/ketall) | Like 'kubectl get all', but _really_ everything | ![GitHub stars](https://img.shields.io/github/stars/corneliusweig/ketall.svg?label=stars&logo=github)
[gke-credentials](https://github.com/danisla/kubefunc) | Fetch credentials for GKE clusters | ![GitHub stars](https://img.shields.io/github/stars/danisla/kubefunc.svg?label=stars&logo=github)
[grep](https://github.com/guessi/kubectl-grep) | Filter Kubernetes resources by matching their names | ![GitHub stars](https://img.shields.io/github/stars/guessi/kubectl-grep.svg?label=stars&logo=github)
[iexec](https://github.com/gabeduke/kubectl-iexec) | Interactive selection tool for `kubectl exec` | ![GitHub stars](https://img.shields.io/github/stars/gabeduke/kubectl-iexec.svg?label=stars&logo=github)
[ingress-nginx](https://kubernetes.github.io/ingress-nginx/kubectl-plugin/) | Interact with ingress-nginx | ![GitHub stars](https://img.shields.io/github/stars/kubernetes/ingress-nginx.svg?label=stars&logo=github)
[konfig](https://github.com/corneliusweig/konfig) | Merge, split or import kubeconfig files | ![GitHub stars](https://img.shields.io/github/stars/corneliusweig/konfig.svg?label=stars&logo=github)
[krew](https://sigs.k8s.io/krew) | Package manager for kubectl plugins. | ![GitHub stars](https://img.shields.io/github/stars/kubernetes-sigs/krew.svg?label=stars&logo=github)
[kubesec-scan](https://github.com/stefanprodan/kubectl-kubesec) | Scan Kubernetes resources with kubesec.io. | ![GitHub stars](https://img.shields.io/github/stars/stefanprodan/kubectl-kubesec.svg?label=stars&logo=github)
[kudo](https://kudo.dev/) | Declaratively build, install, and run operators using KUDO. | 
[match-name](https://github.com/gerald1248/kubectl-match-name) | Match names of pods and other API objects | ![GitHub stars](https://img.shields.io/github/stars/gerald1248/kubectl-match-name.svg?label=stars&logo=github)
[mtail](https://github.com/ahmetb/kubectl-extras) | Tail logs from multiple pods matching label selector | ![GitHub stars](https://img.shields.io/github/stars/ahmetb/kubectl-extras.svg?label=stars&logo=github)
[node-admin](https://github.com/danisla/kubefunc) | List nodes and run privileged pod with chroot | ![GitHub stars](https://img.shields.io/github/stars/danisla/kubefunc.svg?label=stars&logo=github)
[oidc-login](https://github.com/int128/kubelogin) | Login for OpenID Connect authentication | ![GitHub stars](https://img.shields.io/github/stars/int128/kubelogin.svg?label=stars&logo=github)
[open-svc](https://github.com/superbrothers/kubectl-open-svc-plugin) | Open the Kubernetes URL(s) for the specified service in your browser. | ![GitHub stars](https://img.shields.io/github/stars/superbrothers/kubectl-open-svc-plugin.svg?label=stars&logo=github)
[pod-logs](https://github.com/danisla/kubefunc) | Display a list of pods to get logs from | ![GitHub stars](https://img.shields.io/github/stars/danisla/kubefunc.svg?label=stars&logo=github)
[pod-shell](https://github.com/danisla/kubefunc) | Display a list of pods to execute a shell in | ![GitHub stars](https://img.shields.io/github/stars/danisla/kubefunc.svg?label=stars&logo=github)
[preflight](https://github.com/replicatedhq/troubleshoot) | Executes application preflight tests in a cluster | ![GitHub stars](https://img.shields.io/github/stars/replicatedhq/troubleshoot.svg?label=stars&logo=github)
[prompt](https://github.com/jordanwilson230/kubectl-plugins/tree/krew#kubectl-prompt) | Prompts for user confirmation when executing commands in critical namespaces or clusters, i.e., production. | ![GitHub stars](https://img.shields.io/github/stars/jordanwilson230/kubectl-plugins.svg?label=stars&logo=github)
[prune-unused](https://github.com/thecloudnatives/kubectl-plugins) | Prune unused resources | ![GitHub stars](https://img.shields.io/github/stars/thecloudnatives/kubectl-plugins.svg?label=stars&logo=github)
[rbac-lookup](https://github.com/reactiveops/rbac-lookup) | Reverse lookup for RBAC | ![GitHub stars](https://img.shields.io/github/stars/reactiveops/rbac-lookup.svg?label=stars&logo=github)
[rbac-view](https://github.com/jasonrichardsmith/rbac-view) | A tool to visualize your RBAC permissions. | ![GitHub stars](https://img.shields.io/github/stars/jasonrichardsmith/rbac-view.svg?label=stars&logo=github)
[resource-capacity](https://github.com/robscott/kube-capacity) | Provides an overview of resource requests, limits, and utilization | ![GitHub stars](https://img.shields.io/github/stars/robscott/kube-capacity.svg?label=stars&logo=github)
[restart](https://github.com/achanda/kubectl-restart) | Restarts a pod with the given name | ![GitHub stars](https://img.shields.io/github/stars/achanda/kubectl-restart.svg?label=stars&logo=github)
[rm-standalone-pods](https://github.com/ahmetb/kubectl-extras) | Remove all pods without owner references | ![GitHub stars](https://img.shields.io/github/stars/ahmetb/kubectl-extras.svg?label=stars&logo=github)
[sniff](https://github.com/eldadru/ksniff) | easly start a remote packet capture on kubernetes pods using tcpdump and wireshark | ![GitHub stars](https://img.shields.io/github/stars/eldadru/ksniff.svg?label=stars&logo=github)
[sort-manifests](https://github.com/superbrothers/ksort) | Sort manfest files in a proper order by Kind | ![GitHub stars](https://img.shields.io/github/stars/superbrothers/ksort.svg?label=stars&logo=github)
[ssh-jump](https://github.com/yokawasa/kubectl-plugin-ssh-jump) | A kubectl plugin to SSH into Kubernetes nodes using a SSH jump host Pod | ![GitHub stars](https://img.shields.io/github/stars/yokawasa/kubectl-plugin-ssh-jump.svg?label=stars&logo=github)
[sudo](https://github.com/postfinance/kubectl-sudo) | Run Kubernetes commands impersonated as group system:masters | ![GitHub stars](https://img.shields.io/github/stars/postfinance/kubectl-sudo.svg?label=stars&logo=github)
[tail](https://github.com/boz/kail) | Stream logs from multiple pods and containers using simple, dynamic source selection. | ![GitHub stars](https://img.shields.io/github/stars/boz/kail.svg?label=stars&logo=github)
[view-secret](https://github.com/ahmetb/kubectl-extras) | Decode secrets | ![GitHub stars](https://img.shields.io/github/stars/ahmetb/kubectl-extras.svg?label=stars&logo=github)
[view-serviceaccount-kubeconfig](https://github.com/superbrothers/kubectl-view-serviceaccount-kubeconfig-plugin) | Show a kubeconfig setting to access the apiserver with a specified serviceaccount. | ![GitHub stars](https://img.shields.io/github/stars/superbrothers/kubectl-view-serviceaccount-kubeconfig-plugin.svg?label=stars&logo=github)
[view-utilization](https://github.com/etopeter/kubectl-view-utilization) | Shows cluster cpu and memory utilization | ![GitHub stars](https://img.shields.io/github/stars/etopeter/kubectl-view-utilization.svg?label=stars&logo=github)
[virt](https://kubevirt.io) | Control KubeVirt virtual machines using virtctl | 
[warp](https://github.com/ernoaapa/kubectl-warp) | Sync and execute local files in Pod | ![GitHub stars](https://img.shields.io/github/stars/ernoaapa/kubectl-warp.svg?label=stars&logo=github)
[who-can](https://github.com/aquasecurity/kubectl-who-can) | Shows who has RBAC permissions to access Kubernetes resources | ![GitHub stars](https://img.shields.io/github/stars/aquasecurity/kubectl-who-can.svg?label=stars&logo=github)


---

_This page is generated by running the
[generate-plugin-overview](http://sigs.k8s.io/krew/cmd/generate-plugin-overview)
tool._

