# ArgoCD
Forked from: [argoproj/argo-cd](https://github.com/argoproj/argo-cd)

## v2.4.2 (Base)
Argo CD's latest stable release, as of June 25, 2022, is v2.4.2. The list of Argo CD releases can be accessed [here](https://github.com/argoproj/argo-cd/releases)

## v2.3.5 (Fork)
Argo CD has breaking changes for plugins for v2.4.x:

>Update plugins to use newly-prefixed environment variables
If you use plugins that depend on user-supplied environment variables, then they must be updated to be compatible with Argo CD 2.4. Here is an example of user-supplied environment variables in the plugin section of an Application spec:

```
apiVersion: argoproj.io/v1alpha1
kind: Application
spec:
  source:
    plugin:
      env:
        - name: FOO
          value: bar
Going forward, all user-supplied environment variables will be prefixed with ARGOCD_ENV_ before being sent to the plugin's init, generate, or discover commands. This prevents users from setting potentially-sensitive environment variables.
```

>If you have written a custom plugin which handles user-provided environment variables, update it to handle the new prefix.

>If you use a third-party plugin which does not explicitly advertise Argo CD 2.4 support, it might not handle the prefixed environment variables. Open an issue with the plugin's authors and confirm support before upgrading to Argo CD 2.4.

The above means that none of the applications will be able to use a user-defined backend service because the Argo CD Vault Plugin currently doesn't provide support to understand the prefixes. 

The [release post](https://blog.argoproj.io/breaking-changes-in-argo-cd-2-4-29e3c2ac30c9) mentions the following:

> We'll continue publishing security patches for 2.3.x until 2.6.0 is released.

Because of the above, we proceeded to use v2.3.5, which is the latest 2.3.x version.


### Resource Customization ConfigMap
Pulls in resource overrides from the resource customization `ConfigMap`. This `ConfigMap` will only exist if created by 
ArgoCD Extensions. If this ConfigMap doesn't exist, then ArgoCD simply continues on.

The resource overrides are loaded ***after*** the `argocd-cm` `ConfigMap`. If a resource override was already provided in the `argocd-cm` `ConfigMap`, then the resource override from the resource customization `ConfigMap` is ***not loaded***.