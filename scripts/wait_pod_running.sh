#!/bin/bash

set -euo pipefail

readonly NAME=$1

while :
do
  status=$(kubectl --context=kind-kind get pods -o json | jq -r ".items[] | select(.metadata.name | contains(\"${NAME}\")) | .status.phase")
  if [[ $status = 'Running' ]]; then
    break
  fi
  echo "Waiting for ${NAME} to be ready..."
  sleep 3
done
