#!/bin/bash

if [[ $# -ne 3 ]]; then
  echo "Usage: ./create-node-lease.sh NODE_NAME NAMESPACE_NAME LEASE_DURATION"
  exit 1
fi

NODE_NAME=$1
NAMESPACE_NAME=$2
LEASE_DURATION=$3
NODE_UID="$(kubectl get node -o custom-columns=uid:.metadata.uid --no-headers $NODE_NAME)"

echo "{\"apiVersion\": \"coordination.k8s.io/v1\",\"kind\": \"Lease\",\"metadata\": {\"name\": \"$NODE_NAME\",\"namespace\": \"$NAMESPACE_NAME\", \"ownerReferences\": [{\"apiVersion\": \"v1\",\"kind\": \"Node\",\"name\": \"$NODE_NAME\",\"uid\": \"$NODE_UID\"}]}, \"spec\": {\"holderIdentity\": \"$NODE_NAME\", \"leaseDurationSeconds\": $LEASE_DURATION, \"renewTime\": \"$(date -u +"%Y-%m-%dT%H:%M:%S.000000Z")\"}}" | kubectl apply -f -
