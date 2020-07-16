# Kommander Addon Repository

This is a [kubernetes addon](https://github.com/mesosphere/kubeaddons) repository which contains the Kommander addon.

# Using and Releasing `kubeaddons-kommander` Repository

- Pre-releases are only tagged with `testing-` tags, no `stable-` tags for pre-releases.
- The `-rc` that will actually be the stable GA release will be used to branch off of `master` to its own `stable-` branch
- Kommander follows a more semver style versioning and is setup to be able to support multiple major and minor versions at the same time.

## Step by step guides to release Kommander versions (pre-releases, latest and previously released)

### Tag/release pre-releases and update SOAK

1. fetch latest repo state: `git fetch`
1. apply tags to the commit you want to tag (usually `HEAD`), kommander uses multiple tags at this point:
   1. apply and push "base" semver based tag: e.g. `git tag v1.1.0-beta.3 && git push origin v1.1.0-beta.3`
   1. apply and push "consumable" testing tag(s) for each supported k8s version: e.g. `git tag testing-1.16-1.1.0-beta.3 && git push origin testing-1.16-1.1.0-beta.3`
1. head to github and update release information for that prerelease: check [releases page](https://github.com/mesosphere/kubeaddons-kommander/releases) for up-to-date example. _Make sure to edit the "base" release_.
1. update SOAK by updating kommander `configVersion` tag in its `cluster.yaml`
1. update `mergebot-config.json` on `master` branch (only there) to reflect the next release (usually just increase the prerelease number)
1. create a new revision file and bump the `appversion.kubeaddons.mesosphere.io/kommander` annotation to reflect the next release (usually just increase the prerelease number)

#### Hotfixing pre-releases when problems are found on SOAK

Instead of adding a hotfix to an already established pre-release we will create a new one.
When a major issue in "beta 3" is found on SOAK, we will follow the steps above to create "beta 4", and update SOAK with that. Same is true for RCs.

### Tag/release latest stable version

Only pre-releases that are SOAKed for at least two weeks should be used as stable releases. After the soaking period was successful, follow these steps:

1. fetch latest repo state: `git fetch`
1. update kommanders `appVersion` to be stable by removing the pre-release suffix and commit that change to `master` (remember to also update revision)
1. apply and push tags
   1. "base" semver based tag: e.g. `git tag v1.1.0 && git push origin v1.1.0`
   1. "consumable" stable tag(s) for supported k8s versions: e.g. `git tag stable-1.16-1.1.0 && git push origin stable-1.16-1.1.0`
1. create new `stable-*` branch for that minor release: e.g. `git checkout -b stable-1.1.x`. that way future updates have an easy target.
1. head to github and update release information for that release: check [releases page](https://github.com/mesosphere/kubeaddons-kommander/releases) for up-to-date example. _Make sure to edit the "base" release_.
1. add that new `stable-*` branch to `mergebot-config.json` on `master` and set its version to the next patch release (usually `.1`)
1. in order to allow backports to that newely reated minor version, make sure that the charts minor version also is bumped. 
1. to make it easy for fellow colleagues, create a new directory (e.g. `1.2.0` for people to work in) 

### Dealing with previously released stable versions

Sometimes we might need to push a fix for an older version, in these cases we need to use the `stable-*` branch for these versions. E.g. in order to be able to push "Kommander 1.0.1" after `master` already is in a WIP "1.1.x" state, we have `stable-1.0.x` branch to release `v1.0.1` tag.

#### Tag/release previously released versions (patch releases)

1. fetch latest repo state: `git fetch`
1. checkout respective stable branch: e.g. `git checkout stable-1.0.x`
1. apply tags:
   1. apply and push "base" semver based tag: e.g. `git tag v1.0.1 && git push origin v1.0.1`
   1. apply and push "consumable" stable tag(s) for each supported k8s version: e.g. `git tag stable-1.16-1.0.1 && git push origin stable-1.16-1.0.1`
1. merge that `stable-*` branch back to `master` in order to keep all information on master
1. update `mergebot-config.json` on `master` and set its version to the next patch release
