# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

# Check if NUMAFLOW_VERSION is set, if not, then echo message about set it
ifndef NUMAFLOW_VERSION
$(error NUMAFLOW_VERSION is not set. Please set it to the version you want to release, for example: v1.4.0)
endif

# Update the numaflow CRDs
.PHONY: upgrade-charts
upgrade-charts:
	cd upgrade && mkdir -p bin && \
	go build -o bin/upgrade main.go && \
	NUMAFLOW_VERSION=${NUMAFLOW_VERSION} ./bin/upgrade
