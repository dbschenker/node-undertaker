#!/bin/bash

helm upgrade --install --create-namespace -n node-undertaker node-undertaker node-undertaker \
  --set deployment.image.tag=local \
  --set deployment.settings.cloudProvider=kind \
  --set reporter.image.tag=latest \
  --set deployment.settings.logLevel=debug \
  --set-string deployment.podAnnotations.prometheus\\.io\\/scrape=true \
  --set deployment.podAnnotations."prometheus\.io\/path"=/metrics \
  --set-string deployment.podAnnotations.prometheus\\.io\\/port=8080 \
  --set deployment.settings.nodeLeaseNamespace=kube-node-lease \
  $@
