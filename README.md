# Kommander Addon Repository

This is a [kubernetes addon](https://github.com/mesosphere/kubeaddons) repository which contains [Kommander](https://github.com/mesosphere/kommander) related addon applications.

# Updating Kommander Addon

**This section is only intended for short term guideance and might not be/stay 100% accurate, please get rid of whatever you see that seems wrong.**
**this repo follows (mostly?) the same rules as https://github.com/mesosphere/kubernetes-base-addons - so please check readme files over there**

to update kommander chart, _rename_ current revision and update the file as needed (usualy just chart version).
in case we need to support multiple versions at the same time, its a good idea to use directories for it (1.x, 2.x, maybe even 1.1.x and 1.2.x if necessary, we can rename directories as needed). supporting multiple revisions at the same time shouldnt be necessary for kommander at this point.
