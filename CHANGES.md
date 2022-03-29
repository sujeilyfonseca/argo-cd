# ArgoCD

Forked from [argoproj/argo-cd](https://github.com/argoproj/argo-cd), commit [987f665](https://github.com/argoproj/argo-cd/tree/987f6659b88e656a8f6f8feef87f4dd467d53c44).

## v2.2.3 (Base)

The following sections detail the enhancements made to ArgoCD Extensions.

## v2.3.2 (Fork)

### Resource Customization ConfigMap

Pulls in resource overrides from the resource customization ConfigMap. This ConfigMap will only exist if created by 
ArgoCD Extensions. If this ConfigMap doesn't exist, then ArgoCD simply continues on.

The resource overrides are loaded _after_ the `argocd-cm` ConfigMap. If a resource override was already provided in the
`argocd-cm` ConfigMap, then the resource override from the resource customization ConfigMap is ***not loaded***.
