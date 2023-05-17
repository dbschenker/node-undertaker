#!/bin/bash

helm template --create-namespace -n node-undertaker node-undertaker node-undertaker $@

