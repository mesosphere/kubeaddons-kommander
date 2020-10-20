# Kommander Addon Repository

This is a [kubernetes addon](https://github.com/mesosphere/kubeaddons) repository which contains the Kommander addon.

# Contributing to this repository

This [KEP](https://github.com/mesosphere/ksphere-platform/blob/master/keps/sig-ksphere-catalog/20200818-remove-revisions.md) changed the way we handle revision, which enables this repository to only maintain one revision per branch and removes the need to keep a flat hierarchy on master.

Kommander currently is supported in three different versions, which are reporesented by their respective branches:

- 1.3 being developed on `master` branch
- 1.2 living on `stable-1.2.x` branch
- 1.1 living on `stable-1.1.x` branch
- 1.0 living on `stable-1.0.x` branch

# Creating a PR against this repo

If you want to bump your chart version or change the addon itself, please do so by changing the newest existing file on the respective branch.
You still need to update the addon revision, but you **do not** need to copy the whole file anymore.

# Releasing from this repo

- Kommander follows a semver style versioning and is set up to be able to support multiple major and minor versions at the same time. (See branches above)

## Step by step guides to release Kommander versions (pre-releases, latest and previously released)

### Tag/release pre-releases on `master` and update SOAK

Cutting a new pre-release is mainly adding a tag to `master` branch and updating some metadata.

1. fetch latest repo state: `git fetch` and make sure you're on `master`
1. apply tags to the commit you want to tag (usually `HEAD`), kommander uses multiple tags at this point:
   1. apply and push "base" semver based tag: e.g. `git tag v1.1.0-beta.3 && git push origin v1.1.0-beta.3`
   1. apply and push "consumable" testing tag(s) for each supported k8s version: e.g. `git tag testing-1.16-1.1.0-beta.3 && git push origin testing-1.16-1.1.0-beta.3`
1. head to github and update release information for that prerelease: check [releases page](https://github.com/mesosphere/kubeaddons-kommander/releases) for up-to-date example.
1. update SOAK by updating kommander `configVersion` tag in its `cluster.yaml`
1. update repo for the next pre-release
   1. update `mergebot-config.json` on `master` branch (only there) to reflect the next release (usually just increase the prerelease number)
   1. create a new revision file (e.g. `kommander-1.1.0-beta.4.yaml`) and bump the `appversion.kubeaddons.mesosphere.io/kommander` annotation to reflect the next release (usually just increase the prerelease number). You do not need to reflect the revision in the filename anymore.

#### Hotfixing pre-releases when problems are found on SOAK

Instead of adding a hotfix to an already established pre-release we will create a new one on `master` branch. Just follow the steps above.
When a major issue in "beta 3" is found on SOAK, we will follow the steps above to create "beta 4", and update SOAK with that. Same is true for RCs.

### Tag/release latest stable version

New GA releases mainly happen on master branch, a `stable-*` maintenance branch is created for future patch releases.

Only pre-releases that are SOAKed for at least two weeks should be used as stable releases. After the soaking period was successful, follow these steps:

1. fetch latest repo state: `git fetch` and make sure you're on `master`
1. update the `revision` and `appVersion` to be stable by removing the pre-release suffix and commit that change to `master` (remember to also update revision)
1. apply and push tags
   1. "base" semver based tag: e.g. `git tag v1.1.0 && git push origin v1.1.0`
   1. "consumable" stable tag(s) for supported k8s versions: e.g. `git tag stable-1.16-1.1.0 && git push origin stable-1.16-1.1.0`
1. create new `stable-*` branch for that minor release: e.g. `git checkout -b stable-1.1.x`. that way future updates have an easy target.
1. head to github and update release information for that release: check [releases page](https://github.com/mesosphere/kubeaddons-kommander/releases) for up-to-date example.
1. add that new `stable-*` branch to `mergebot-config.json` on `master` and set its version to the next patch release (usually `.1`)
1. in order to allow backports to that newly reated minor version, make sure that the charts minor version also is bumped.
1. to make it easy for fellow colleagues, create a new directory on `master` (e.g. `1.2.0`) and your new stable branch (e.g. `1.1.1`) for people to work in)

There is no need to merge back `stable-*` branches into master since we don't need to maintain a flat history anymore.

### Dealing with previously released stable versions

Sometimes we might need to push a fix for an older version, in these cases we need to use the `stable-*` branch for these versions. E.g. in order to be able to push "Kommander 1.0.1" after `master` already is in a WIP "1.1.x" state, we have `stable-1.0.x` branch to release `v1.0.1` tag.

#### Tag/release previously released versions (patch releases)

1. fetch latest repo state: `git fetch`
1. checkout respective stable branch: e.g. `git checkout stable-1.0.x`
1. apply tag: e.g. `git tag v1.0.1 && git push origin v1.0.1`
1. update `mergebot-config.json` on `master` and set its version to the next patch release

There is no need to merge back `stable-*` branches into master since we dont need to maintain a flat history anymore.

# Tests

## E2E

This repo runs the E2E tests against the PRs of this repository.
You can find more information [in this README](https://github.com/mesosphere/kommander/tree/master/system-tests#system-tests).
