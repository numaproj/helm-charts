#!/bin/bash

# This will ensure that the script fails if any command fails
set -euo pipefail

#################### Version update start ####################
# Read appVersion and version from chart.yaml
APP_VERSION=$(grep 'appVersion:' charts/numaflow/Chart.yaml | awk '{print $2}' | sed 's/"//g')
VERSION=$(grep 'version:' charts/numaflow/Chart.yaml | awk '{print $2}')

# Extract the numeric part of the NUMAFLOW_VERSION
NUMAFLOW_NUMERIC_VERSION=${NUMAFLOW_VERSION#v}

# Compare NUMAFLOW_VERSION and APP_VERSION
if [[ "$NUMAFLOW_NUMERIC_VERSION" == "$APP_VERSION" ]]; then
  echo "Versions are identical. No update is necessary."
else
  # Split versions into arrays
  IFS='.' read -r -a numaflow_parts <<< "$NUMAFLOW_NUMERIC_VERSION"
  IFS='.' read -r -a app_parts <<< "$APP_VERSION"
  IFS='.' read -r -a current_parts <<< "$VERSION"

  # Initialize variables to determine the type of update
  major_update=0
  minor_update=0
  patch_update=0

  # Determine if the change is major, minor, or patch
  if [[ "${numaflow_parts[0]}" -ne "${app_parts[0]}" ]]; then
    major_update=1
  elif [[ "${numaflow_parts[1]}" -ne "${app_parts[1]}" ]]; then
    minor_update=1
  elif [[ "${numaflow_parts[2]}" -ne "${app_parts[2]}" ]]; then
    patch_update=1
  fi

  # Perform version update accordingly
  if [[ $major_update -eq 1 ]]; then
    new_version="$((current_parts[0]+1)).0.0"
    echo "Major update detected. Updated version: $new_version"
  elif [[ $minor_update -eq 1 ]]; then
    new_version="${current_parts[0]}.$((current_parts[1]+1)).0"
    echo "Minor update detected. Updated version: $new_version"
  elif [[ $patch_update -eq 1 ]]; then
    new_version="${current_parts[0]}.${current_parts[1]}.$((current_parts[2]+1))"
    echo "Patch update detected. Updated version: $new_version"
  else
    echo "No significant version change detected. No update to $VERSION required."
    exit 1
  fi
  sed -i.bak "s/version: .*/version: ${new_version}/" charts/numaflow/Chart.yaml
  sed -i.bak "s/appVersion: .*/appVersion: \"${NUMAFLOW_NUMERIC_VERSION}\"/" charts/numaflow/Chart.yaml
  echo "Updated version: $new_version and appVersion: $NUMAFLOW_NUMERIC_VERSION in charts/numaflow/Chart.yaml"
fi
#################### Version update end ####################

