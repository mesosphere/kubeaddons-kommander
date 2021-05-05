# Kommander Addon Repository

This is a [kubernetes addon](https://github.com/mesosphere/kubeaddons) repository which contains the Kommander addon.

# Contributing to this repository

This [KEP](https://github.com/mesosphere/ksphere-platform/blob/master/keps/sig-ksphere-catalog/20200818-remove-revisions.md) changed the way we handle revision, which enables this repository to only maintain one revision per branch and removes the need to keep a flat hierarchy on master.

Kommander currently is supported in three different versions, which are represented by their respective branches:

- 1.5 being developed on `master` branch
- 1.4 living on `1.4.x` branch
- 1.3 living on `1.3.x` branch
- 1.2 living on `1.2.x` branch
- 1.1 living on `1.1.x` branch
- 1.0 living on `1.0.x` branch

# Creating a PR against this repo

If you want to bump your chart version or change the addon itself, please do so by changing the newest existing file on the respective branch.
You still need to update the addon revision, but you **do not** need to copy the whole file anymore.

# Releasing from this repo

Kommander follows a semver style versioning and is set up to be able to support multiple major and minor versions at the same time. (See branches above)

## Step by step guides to release Kommander versions (pre-releases, latest and previously released)

### Tag/release pre-releases on `master` branch and update SOAK

Cutting a new pre-release is mainly adding a tag to `master` branch and updating some metadata.

1. fetch latest repo state: `git fetch` and make sure you're on `master`
1. apply tags to the commit you want to tag (usually `HEAD`), kommander uses multiple tags at this point:
   1. apply and push "base" semver based tag: e.g. `git tag v1.1.0-beta.3 && git push origin v1.1.0-beta.3`
   1. apply and push "consumable" testing tag(s) for each supported k8s version: e.g. `git tag testing-1.16-1.1.0-beta.3 && git push origin testing-1.16-1.1.0-beta.3`
1. head to github and update release information for that prerelease: check [releases page](https://github.com/mesosphere/kubeaddons-kommander/releases) for up-to-date example.
   1. at least update the component information for easy reference later on
1. ~update SOAK by updating kommander `configVersion` tag in its `cluster.yaml`~ this should be automated by now
1. update repo for the next pre-release
   1. update `mergebot-config.json` on `master` branch (only there) to reflect the next release (usually just increase the prerelease number)
   1. update addon yaml file and bump the `appversion.kubeaddons.mesosphere.io/kommander` annotation to reflect the next release (usually just increase the prerelease number). This change also needs a revision bump!

#### Hotfixing pre-releases when problems are found on SOAK

Instead of adding a hotfix to an already established pre-release we will create a new one on `master` branch. Just follow the steps above.
When a major issue in "beta 3" is found on SOAK, we will follow the steps above to create "beta 4", and update SOAK with that. Same is true for RCs.

### Prepare Minor GA Release / Branch off maintenance branch

New pre-releases mainly happen on master branch, at some point a `[0-9].[0-9].x` maintenance branch is created to prepare the GA release of a minor version.

1. fetch latest repo state: `git fetch` and make sure you're on `master`
1. create new `[0-9].[0-9].x` branch for that minor release: e.g. `git checkout -b 1.1.x`. that way future updates have an easy target and master can carry on with the next minor version.
1. in order to allow backports to that newly created minor version, make sure that the chart's minor version also is bumped.
1. to make it easy for fellow colleagues, rename the existing directory on `master` (e.g. `1.1` -> `1.2`) and update the addons metadata (appversion & revision)
1. update `mergebot-config.json` on `master` to add a new line for the new stable branch and update the `master` mapping to the next release

From there on, it's very similar to releases from `master` branch, there might be a couple RCs before the actual GA tag is cut. When the time comes to cut the GA release, promote the latest RC tags to GA tags: e.g. `v1.3.0` and corresponding stable tags `stable-1.17-1.3.0`, `stable-1.18-1.3.0`, `stable-1.19-1.3.0`.

There is no need to merge back `[0-9].[0-9].x` branches into master since we don't need to maintain a flat history anymore.

#### Release notes
TBD

### Dealing with previously released stable versions

Sometimes we might need to push a fix for an older version, in these cases we need to use the `[0-9].[0-9].x` branch for these versions. E.g. in order to be able to push "Kommander 1.0.1" after `master` already is in a WIP "1.1.x" state, we have `1.0.x` branch to release `v1.0.1` tag.

#### Tag/release previously released versions (patch releases)

1. fetch latest repo state: `git fetch`
1. checkout respective stable branch: e.g. `git checkout 1.0.x`
1. apply tags: e.g. `git tag v1.0.1 && git push origin v1.0.1` along with corresponding `stable` tags
1. update `mergebot-config.json` on `master` and set its version to the next patch release

There is no need to merge back `[0-9].[0-9].x` branches into master since we don't need to maintain a flat history anymore.
