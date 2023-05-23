#!/bin/bash

go install sigs.k8s.io/kind@latest

# Create the test cluster (the tests also do this) but also set local kube
# config
kind create cluster --name local-tests
kind export kubeconfig --name local-tests

# Install k9s
curl -sS https://webinstall.dev/k9s | bash
