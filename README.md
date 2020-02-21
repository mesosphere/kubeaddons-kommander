# Kommander Addon Repository

This is a [kubernetes addon](https://github.com/mesosphere/kubeaddons) repository which contains [Kommander](https://github.com/mesosphere/kommander) related addon applications.

# Updating Kommander Addon

**This section is only intended for short term guideance and might not be/stay 100% accurate, please get rid of whatever you see that seems wrong.**
**this repo follows the same rules as https://github.com/mesosphere/kubernetes-base-addons - so please check [readme](https://github.com/mesosphere/kubernetes-base-addons/blob/master/README.md) and [contributing](https://github.com/mesosphere/kubernetes-base-addons/blob/master/CONTRIBUTING.md) files over there**

Given we only want to support one revision per kommander version at this point, we can keep it simple and use a single revision file named `kommander.yaml` (no revision in filename). This helps keeping working links and reduces overall maintenance burden.

To update kommander addon, update the file as needed (usualy just chart version) and increase revision in `http://catalog.kubeaddons.mesosphere.io/addon-revision` annotation.
In case we need to support multiple _versions_ at the same time, its a good idea to use directories for these (1.x, 2.x, maybe even 1.1.x and 1.2.x if necessary, we can rename directories as needed). supporting multiple _revisions_ per version at the same time shouldnt be necessary for kommander at this point.

# Releasing Kommander Addon

**same caveat as above, release process is being defined in https://github.com/mesosphere/kubernetes-base-addons/blob/master/RELEASE.md currently.**

current thinking is to have branches per supported kommander version (1.x, 2.x,…) and release tags that follow a k8s based scheme `<stable|beta>-<k8s-version>-<kommander-version>` (e.g. `stable-1.16-1.0.0`). this means for every release of kommander, there are (usually) three release tags (we support k8s n-2 versions).
