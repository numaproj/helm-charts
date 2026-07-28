# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

# Check if NUMAFLOW_VERSION is set, if not, then echo message about set it
ifndef NUMAFLOW_VERSION
$(error NUMAFLOW_VERSION is not set. Please set it to the version you want to release, for example: v1.4.0)
endif

# Build the upgrade binary. Used as a prerequisite by the workflow targets so
# they share a single compiled artifact under upgrade/bin/.
.PHONY: build
build:
	cd upgrade && mkdir -p bin && \
	go build -o bin/upgrade .

# Refresh CRDs, RBAC, ServiceAccounts, Chart.yaml, and values.yaml from the
# numaflow upstream. Existing flow, unchanged behaviour.
.PHONY: upgrade-charts
upgrade-charts: build
	NUMAFLOW_VERSION=${NUMAFLOW_VERSION} ./upgrade/bin/upgrade upgrade-charts

# Mirror the 15 manually-mirrored chart files (configmaps, deployments,
# secrets, services) from upstream numaflow. Defaults to dry-run; pass
# SYNC_FLAGS=--apply once you have reviewed the merge output to write it back.
#
# Examples:
#   NUMAFLOW_VERSION=v1.8.2 make sync-mirrored-files
#   NUMAFLOW_VERSION=v1.8.2 make sync-mirrored-files SYNC_FLAGS=--apply
#   NUMAFLOW_VERSION=v1.8.2 make sync-mirrored-files SYNC_FLAGS="--from-version=v1.8.0 --apply"
.PHONY: sync-mirrored-files
sync-mirrored-files: build
	NUMAFLOW_VERSION=${NUMAFLOW_VERSION} ./upgrade/bin/upgrade sync $(SYNC_FLAGS)
