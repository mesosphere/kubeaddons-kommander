KONVOY_SOURCE_DIR       := .konvoy/source
KONVOY_SOURCE_VERSION   := master
KONVOY_REPOSITORY       := https://github.com/mesosphere/konvoy
KONVOY_SOURCE           := $(KONVOY_SOURCE_DIR)/$(KONVOY_SOURCE_VERSION)
$(KONVOY_SOURCE): gitauth
	mkdir -p $(KONVOY_SOURCE)
ifeq (,$(wildcard $(KONVOY_SOURCE)))
	git clone -b $(KONVOY_SOURCE_VERSION) $(KONVOY_REPOSITORY) $(KONVOY_SOURCE)
else
	cd $(KONVOY_SOURCE) && \
		git fetch origin $(KONVOY_SOURCE_VERSION) && \
		git checkout $(KONVOY_SOURCE_VERSION) && \
		git reset --hard origin/$(KONVOY_SOURCE_VERSION) && \
		git clean -fdx
endif

BUILD_DIR ?= build

auto-provisioning.prepare-chart: gitauth $(KONVOY_SOURCE)
	mkdir -p $(BUILD_DIR)
	$(MAKE) -C $(KONVOY_SOURCE) copy-charts BUILD_OUTPUT=$(PWD)/build/konvoy

KUBEADDONS_SOURCE_DIR       := .kubeaddons/source
KUBEADDONS_SOURCE_VERSION   := master
KUBEADDONS_REPOSITORY       := https://github.com/mesosphere/kubeaddons
KUBEADDONS_SOURCE           := $(KUBEADDONS_SOURCE_DIR)/$(KUBEADDONS_SOURCE_VERSION)
$(KUBEADDONS_SOURCE): gitauth
	mkdir -p $(KUBEADDONS_SOURCE)
ifeq (,$(wildcard $(KUBEADDONS_SOURCE)))
	git clone -b $(KUBEADDONS_SOURCE_VERSION) $(KUBEADDONS_REPOSITORY) $(KUBEADDONS_SOURCE)
else
	cd $(KUBEADDONS_SOURCE) && \
		git fetch origin $(KUBEADDONS_SOURCE_VERSION) && \
		git checkout $(KUBEADDONS_SOURCE_VERSION) && \
		git reset --hard origin/$(KUBEADDONS_SOURCE_VERSION) && \
		git clean -fdx
endif

kubeaddons.prepare-chart: gitauth $(KUBEADDONS_SOURCE)
	mkdir -p $(BUILD_DIR)
	$(MAKE) -C $(KUBEADDONS_SOURCE) copy-charts BUILD_OUTPUT=$(PWD)/build/kubeaddons

.PHONY: gitauth
gitauth:
ifeq ($(CI),true)
	git config --global url.git@github.com:.insteadOf https://github.com/
endif
