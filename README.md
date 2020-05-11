# Kommander Addon Repository

This is a [kubernetes addon](https://github.com/mesosphere/kubeaddons) repository which contains [Kommander](https://github.com/mesosphere/kommander) related addon applications.

# Using and Releasing `kubeaddons-kommander` Repository

This repository mostly follows the same rules that [`kubernetes-base-addons`](https://github.com/mesosphere/kubernetes-base-addons) does. Please familiarize yourself with that documentation, and then see the differences below.

- [README.md](https://github.com/mesosphere/kubernetes-base-addons/blob/master/README.md)
- [RELEASE.md](https://github.com/mesosphere/kubernetes-base-addons/blob/master/RELEASE.md)
- [CONTRIBUTING.md](https://github.com/mesosphere/kubernetes-base-addons/blob/master/CONTRIBUTING.md)

## Deviations from `kubernetes-base-addons` release process

Kommander being a slightly different type of repository than `kubernetes-base-addons`, some details in releasing kommander are different.

- Pre-releases are only tagged with `testing-` tags, no `stable-` tags for pre-releases. The `-rc` that will actually be the stable GA release will be _also_ tagged with `stable-` tag (and w/o `-rc` suffix)
- Kommander follows a more semver style versioning and is setup to be able to support multiple major and minor versions at the same time.

## Step by step guides to release Kommander versions (pre-releases, latest and previously released)

Only certain people are allowed to push to `testing` and `stable` branches. If you are asked to do a release, but dont have permissions to push, please ask your manager.

### Tag/release pre-releases and update SOAK

1. fetch latest repo state: `git fetch`
1. checkout `testing` branch: `git checkout testing`
1. merge `origin/master` into `testing` and push updated `testing` branch: `git merge origin/master && git push`
1. apply tags, kommander uses multiple tags at this point:
   1. apply and push "base" semver based tag: e.g. `git tag v1.1.0-beta.3 && git push origin v1.1.0-beta.3`
   1. apply and push "consumable" testing tag(s) for each supported k8s version: e.g. `git tag testing-1.16-1.1.0-beta.3 && git push origin testing-1.16-1.1.0-beta.3`
1. head to github and update release information for that prerelease: check [releases page](https://github.com/mesosphere/kubeaddons-kommander/releases) for up-to-date example. _Make sure to edit the "base" release_.
1. make sure SOAK gets updated, you can sync with people who tag `kubernetes-base-addons`.

#### Hotfixing testing releases when problems are found on SOAK

Instead of adding a hotfix to an already established release we should create a new one.
When a major issue in "beta 3" is found on SOAK, we should follow the steps above to create "beta 4", and update SOAK with that. Same is true for RCs.

In case we absolutely need to hotfix soak, we need to make sure to also push the fix to `master`:

1. Create PR against `testing` branch _and_ forwardport the fix to `master` too.
1. Update SOAK.

### Tag/release latest stable version

Stable releases should be SOAKed for at least two weeks, after the soaking period was successful, they can be publicly released.

1. fetch latest repo state: `git fetch`
1. checkout `stable` branch: `git checkout stable`
1. _merge_ `origin/testing` into `stable` and push updated `stable` branch: `git merge origin/testing && git push`
1. apply and push "consumable" stable tag(s) for supported k8s versions: e.g. `git tag stable-1.16-1.1.0 && git push origin stable-1.16-1.1.0`
1. head to github and update release information for that prerelease: check [releases page](https://github.com/mesosphere/kubeaddons-kommander/releases) for up-to-date example. _Make sure to edit the "base" release_.

### Dealing with previously released stable versions


Sometimes we might need to push a fix for an older version, in these cases we need to create a dedicated stable branch for these versions. E.g. in order to be able to push "Kommander 1.0.1" after "1.1.x" already was released, we have to create `stable-1.0.x` branch from `v1.0.0` tag. (`git checkout v1.0.0 && git checkout -b stable-1.0.x`). after that happened, we can create PRs against that branch and get things fixed.

#### Tag/release previously released versions

1. fetch latest repo state: `git fetch`
1. checkout respective stable branch: e.g. `git checkout stable-1.0.x`
1. apply tags:
   1. apply and push "base" semver based tag: e.g. `git tag v1.0.1 && git push origin v1.0.1`
   1. apply and push "consumable" stable tag(s) for each supported k8s version: e.g. `git tag stable-1.16-1.0.1 && git push origin stable-1.16-1.0.1`
