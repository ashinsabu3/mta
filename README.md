# mta (Migrate to Argo CD)

The `mta` cli will export Flux components to Argo CD/Harness Gitops agent consumable
CRs. This can be used in order to help migrating from Flux to Argo
CD. 

This is currently in an alpha POC state and doesn't support all possible variations.

# Installation

Install the `mta` binary from the [releases page](https://github.com/ashinsabu3/mta/releases) in your `$PATH`.

There is shell completion for convenience. 

> *NOTE* it's probably `zsh` on a Mac

```shell
mta completion bash
```


## Manual Migration

> *NOTE*: See the [Flux Documentation](https://fluxcd.io/flux/get-started/) for more information about Flux.

After downloading the binary, you can scan your system for `HelmReleases` and `Kustomizations`. Example:

```shell
$ mta scan 
┌───────────────┬─────────────┬─────────────┬─────────────────────────────────────────────────────────────────┐
│ KIND          │ NAME        │ NAMESPACE   │ STATUS                                                          │
├───────────────┼─────────────┼─────────────┼─────────────────────────────────────────────────────────────────┤
│ HelmRelease   │ podinfo     │ flux-system │ Release reconciliation succeeded                                │
│ HelmRelease   │ sample      │ flux-system │ Release reconciliation succeeded                                │
├───────────────┼─────────────┼─────────────┼─────────────────────────────────────────────────────────────────┤
│ Kustomization │ flux-system │ flux-system │ Applied revision: main/f35c47113103d67b20859a2301fa5c88a8f7c6c9 │
└───────────────┴─────────────┴─────────────┴─────────────────────────────────────────────────────────────────┘
```

You can then, migrate them over; for example to migrate the `HelmRelease` called `sample` (in my above example), you can do:

```shell
$ mta helmrelease --name sample
```

You'll see the Argo CD Application that will be created:

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: sample
  namespace: argocd
spec:
  destination:
    namespace: quarkus
    server: https://kubernetes.default.svc
  project: default
  source:
    chart: quarkus
    helm:
      values: |
        build:
          enabled: false
        deploy:
          route:
            enabled: false
        image:
          name: quay.io/ablock/gitops-helm-quarkus
    repoURL: https://redhat-developer.github.io/redhat-helm-charts
    targetRevision: 0.0.3
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
    syncOptions:
    - CreateNamespace=true
    - Validate=false
```

You can pipe this into `kubectl apply` or you can have `mta` do it for you

> *NOTE* You'll have to install Argo CD before running this command

```shell
$ mta helmrelease --name sample --confirm-migrate
```

The same can be done for `Kustomizations`, example:

> *NOTE* `Kustomizations`, because of the nature of how they are setup, are migrated via an ApplicationSet

```shell
$ mta kustomization --name flux-system --confirm-migrate
```

By default, the ApplicationSet created from the `Kustomiation` will exclude the `flux-system` directory. You can exclude other directories that have Flux specific Kubernetes objects by passing the `--exclude-dirs` option.

```shell
$ mta kustomization --name flux-system --exclude-dirs flux-system-extras --confirm-migrate
```

> *NOTE* To exclude more directories, you an pass a comma separated list to `--exclude-dirs`. Example: `--exclude-dirs foo,bar,bazz`. You can also pass `--exclude-dirs` to the `scan` command as well.

For a detailed list of examples, read the [examples](./examples) docs.

## Auto Migration

You can have the `scan` subcommand automatically migrate everything for you

> :bangbang: *NOTE* This option also _deletes_ Flux from the system. Use with caution

```shell
$ mta scan --auto-migrate
```
