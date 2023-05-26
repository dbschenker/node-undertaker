#!/bin/bash

helm upgrade --install --create-namespace -n node-undertaker node-undertaker node-undertaker \
  --set controller.image.tag=local \
  --set controller.settings.cloudProvider=kind \
  --set controller.settings.logLevel=debug \
  --set-string controller.podAnnotations.prometheus\\.io\\/scrape=true \
  --set controller.podAnnotations."prometheus\.io\/path"=/metrics \
  --set-string controller.podAnnotations.prometheus\\.io\\/port=8080 \
  --set controller.settings.nodeLeaseNamespace=kube-node-lease \
  --set controller.settings.nodeSelector="" \
  $@
