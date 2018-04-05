#!/usr/bin/env sh

# order of arguments: KUBECONFIG (base64 encoded to avoid any weird escaping issues)
if [ $# -gt 0 ]; then
  kubeconfig=$1
  if [ ! -z "${kubeconfig}" ]; then
    mkdir -p ~/.kube
    echo ${kubeconfig} | base64 -d > ~/.kube/config
  else
    echo "kubeconfig var empty, saving nothing to ~/.kube/config"
    exit 1
  fi
else
    echo "no arguments were passed in"
    exit 1
fi