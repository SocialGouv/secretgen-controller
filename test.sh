#!/usr/bin/env bash
set -e -x -u

# minikube start --driver=docker

mkdir -p /tmp/bin
export PATH=/tmp/bin:$PATH

eval $(minikube docker-env --shell=bash)
export SECRETGEN_E2E_NAMESPACE=secretgen-test

./hack/build.sh
./hack/deploy.sh

./hack/test.sh # unit tests

# ./hack/test-e2e.sh
go clean -testcache
# create ns if not exists because the `apply -f -` won't complain on a no-op if the ns already exists.
kubectl create ns $SECRETGEN_E2E_NAMESPACE --dry-run=client -o yaml | kubectl apply -f -
go test ./test/e2e/ -run="ExportSuccessful" -timeout 60m -test.v $@

./hack/test-examples.sh

echo ALL SUCCESS