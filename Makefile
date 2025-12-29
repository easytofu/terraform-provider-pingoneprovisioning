# Makefile
PROVIDER_NAME := pingone-propagation
NAMESPACE := easygo
VERSION ?= 0.1.0

.PHONY: publish clean

publish:
	@chmod +x scripts/publish-oci.sh
	@VERSION=$(VERSION) PROVIDER_NAME=$(PROVIDER_NAME) NAMESPACE=$(NAMESPACE) ./scripts/publish-oci.sh

clean:
	rm -rf dist oci-layout
