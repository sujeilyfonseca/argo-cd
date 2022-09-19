# ArgoCD
Forked from: [argoproj/argo-cd](https://github.com/argoproj/argo-cd)

## v2.4.12 (Base)
Argo CD's latest stable release, as of September 16, 2022, is v2.4.12. The list of Argo CD releases can be accessed [here](https://github.com/argoproj/argo-cd/releases)

## v2.4.12-patched (Fork)
The changes were rebased based on v2.4.12. The following section details the enhancements made to Argo CD Extensions that were integrated into Argo CD.

### Resource Customization ConfigMap
Pulls in resource overrides from the resource customization `ConfigMap`. This `ConfigMap` will only exist if created by 
ArgoCD Extensions. If this ConfigMap doesn't exist, then ArgoCD simply continues on.

The resource overrides are loaded ***after*** the `argocd-cm` `ConfigMap`. If a resource override was already provided in the `argocd-cm` `ConfigMap`, then the resource override from the resource customization `ConfigMap` is ***not loaded***.

### Security Patches
We have addressed/are addressing security vulnerabilities associated with ASoC, Twistlock, and Whitesource security scans.