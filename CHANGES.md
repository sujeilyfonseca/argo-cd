# ArgoCD
Forked from: [argoproj/argo-cd](https://github.com/argoproj/argo-cd)

## v2.3.3 (Base)
Argo CD's latest stable release, as of  May 13, 2022, is v2.3.3.

## v2.3.3 (Fork)
The changes were rebased based on v2.3.3. This section details the enhancements made to Argo CD Extensions.

### Resource Customization ConfigMap
Pulls in resource overrides from the resource customization `ConfigMap`. This `ConfigMap` will only exist if created by 
ArgoCD Extensions. If this ConfigMap doesn't exist, then ArgoCD simply continues on.

The resource overrides are loaded ***after*** the `argocd-cm` `ConfigMap`. If a resource override was already provided in the `argocd-cm` `ConfigMap`, then the resource override from the resource customization `ConfigMap` is ***not loaded***.