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
	$(MAKE) -C $(KONVOY_SOURCE) copy-charts BUILD_OUTPUT=$(PWD)/build

.PHONY: gitauth
gitauth:
ifeq ($(CI),true)
	git config --global url.git@github.com:.insteadOf https://github.com/
endif
