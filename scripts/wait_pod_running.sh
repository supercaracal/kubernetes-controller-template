#!/bin/bash

set -euo pipefail

while :
do
  status=$(kubectl --context=kind-kind get pods -o json | jq -r ".items[] | select(.metadata.name | contains('$1')) | .status.phase")
  if [[ $status = 'Running' ]]; then
    break
  fi
  echo "Waiting for $1 to be ready..."
  sleep 3
done
