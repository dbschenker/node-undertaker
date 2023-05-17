#!/bin/bash

helm upgrade --install --create-namespace -n node-undertaker node-undertaker node-undertaker --set deployment.image.tag=local --set reporter.image.tag=latest

