#!/usr/bin/env bash

################################################################################
# Following assumptions are made by this script:                               #
# * gcloud, docker, and docker-compose is installed                            #
# * it is called from $GOPATH/src/github.com/google/keytransparency            #
# * there is a project called key-transparency on gce which has a cluster      #
#   called "ci-cluster" within the "us-central1-a" compute zone                #
# * there is a service key to authenticate with glcoud to above project in     #
#   $GOPATH/src/github.com/google/keytransparency/service_key.json             #
################################################################################

PROJECT_NAME=key-transparency
NAME_SPACE=default

MAX_RETRY=30

function main()
{
  # create key-pairs:
  ./scripts/prepare_server.sh -f
  initGcloud
  buildDockerImgs
  tearDown
  pushTrillianImgs

  # Deploy all trillian related services:
  kubectl apply -f deploy/kubernetes/trillian-deployment.yml

  pushKTImgs
  waitForTrillian
  createTreeAndSetIDs

  # Deploy all keytransparency related services (server and signer):
  kubectl apply -f deploy/kubernetes/keytransparency-deployment.yml
}

function initGcloud()
{
  gcloud config set project ${PROJECT_NAME}
  gcloud config set compute/zone us-central1-a
  gcloud container clusters get-credentials ci-cluster
}

function buildDockerImgs()
{
  # Work around some git permission issues on linux:
  chmod a+r ../trillian/storage/mysql/storage.sql

  # Build all images defined in the docker-compose.yml:
  docker-compose build
}

function pushTrillianImgs()
{
  gcloud docker -- push us.gcr.io/${PROJECT_NAME}/db
  images=("db" "trillian_log_server" "trillian_map_server" "trillian_log_signer")
  for DOCKER_IMAGE_NAME in "${images[@]}"
  do
    # Push the images as we refer to them in the kubernetes config files:
    gcloud docker -- push us.gcr.io/${PROJECT_NAME}/${DOCKER_IMAGE_NAME}
  done
}

function pushKTImgs()
{
  images=("keytransparency-server" "keytransparency-signer" "prometheus")
  for DOCKER_IMAGE_NAME in "${images[@]}"
  do
    # Push the images as we refer to them in the kubernetes config files:
    gcloud docker -- push us.gcr.io/${PROJECT_NAME}/${DOCKER_IMAGE_NAME}
  done
}

function waitForTrillian()
{
  # It's very unlikely that everything is up running before 15 sec.:
  sleep 15
  # Wait for trillian-map pod to be up:
  COUNTER=0
  MAPSRV=""
  until [ -n "$MAPSRV" ] || [  $COUNTER -gt $MAX_RETRY ]; do
    # Service wasn't up yet:
    sleep 10;
    let COUNTER+=1
    MAPSRV=$(kubectl get pods --selector=run=trillian-map -o jsonpath={.items[*].metadata.name});
  done

  if [ -n "$MAPSRV" ]; then
    echo "trillian-map service is up"
  else
    echo "Stopped waiting for trillian-map service. Quitting ..."
    exit 1;
  fi
}

function createTreeAndSetIDs()
{
  LOG_ID=""
  MAP_ID=""
  COUNTER=0
  until [ -n "$LOG_ID" ] || [  $COUNTER -gt $MAX_RETRY ]; do
    # RPC was not available yet, wait and retry:
    sleep 10;
    let COUNTER+=1
    # get the currentl trillian-map pod:
    MAPSRV=$(kubectl get pods --selector=run=trillian-map -o jsonpath={.items[*].metadata.name});
    LOG_ID=$(echo 'go run $GOPATH/src/github.com/google/trillian/cmd/createtree/main.go --admin_server=localhost:8090 --pem_key_path=testdata/log-rpc-server.privkey.pem --pem_key_password="towel" --signature_algorithm=ECDSA --tree_type=LOG' | kubectl exec -i $MAPSRV -- /bin/sh )
    MAP_ID=$(echo 'go run $GOPATH/src/github.com/google/trillian/cmd/createtree/main.go --admin_server=localhost:8090 --pem_key_path=testdata/log-rpc-server.privkey.pem --pem_key_password="towel" --signature_algorithm=ECDSA --tree_type=MAP' | kubectl exec -i $MAPSRV -- /bin/sh )
  done

  if [ -n "$LOG_ID" ] && [ -n "$MAP_ID" ]; then
    echo "Trees created with MAP_ID=$MAP_ID and LOG_ID=$LOG_ID"
    # Substitute LOG_ID and MAP_ID in template kubernetes file:
    sed 's/${LOG_ID}'/${LOG_ID}/g deploy/kubernetes/keytransparency-deployment.yml.tmpl > deploy/kubernetes/keytransparency-deployment.yml
    sed -i 's/${MAP_ID}'/${MAP_ID}/g deploy/kubernetes/keytransparency-deployment.yml
  else
    echo "Failed to create tree. Need map-id and log-id before running kt-server/-signer."
    exit 1
  fi
}

function tearDown()
{
  kubectl delete --all services --namespace=$NAME_SPACE
  kubectl delete --all deployments --namespace=$NAME_SPACE
  kubectl delete --all pods --namespace=$NAME_SPACE
}

# Run everything:
main
