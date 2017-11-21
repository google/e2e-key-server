#!/usr/bin/env bash

################################################################################
# Following assumptions are made by this script:                               #
# * gcloud, docker, and docker-compose is installed                            #
# * it is called from $GOPATH/src/github.com/google/keytransparency            #
# * there is a project called key-transparency on gce  which has has gke       #
#   enabled and a cluster configured; gcloud is already set to this            #
#   project via:                                                               #
#   # see gcloud help auth and authenticate, then:                             #
#   gcloud config set project key-transparency                                 #
#   gcloud container clusters get-credentials <your-cluster-name>              #
#   gcloud config set compute/zone <your-compute-zone>                         #
#                                                                              #
#   See the project's .travis.yml file for a working example.                  #
#                                                                              #
################################################################################

set -ex
PROJECT_NAME=key-transparency

MAX_RETRY=30

dir="$(dirname "$0")"
# Import createTree commands:
source ${dir}/configure_trillian.sh

function main()
{
  checkCmdsAvailable
  # create key-pairs:
  ./scripts/prepare_server.sh -f
  prepareSecrets
  buildDockerImgs

  # Deploy all trillian related services.
  # the following line only restarts the DB if its config changed:
  kubectl apply -f deploy/kubernetes/db-deployment.yml
  kubectl apply -f deploy/kubernetes/trillian-deployment.yml
  # Deploy all keytransparency related services (server and signer):
  kubectl apply -f deploy/kubernetes/keytransparency-deployment.yml

  waitForTrillian
  createTreeIfNeccessaryAndSetIDs
}


function buildDockerImgs()
{
  # Work around some git permission issues on linux:
  chmod a+r ../trillian/storage/mysql/storage.sql

  # Build all images defined in the docker-compose.yml:
  docker-compose build
  docker-compose push
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
    MAPSRV=$(kubectl get pods --selector=run=map_server -o jsonpath={.items[*].metadata.name});
  done

  if [ -n "$MAPSRV" ]; then
    echo "trillian-map service is up"
  else
    echo "Stopped waiting for trillian-map service. Quitting ..."
    exit 1;
  fi
}

function createTreeIfNeccessaryAndSetIDs()
{
  LOG_ID=""
  MAP_ID=""
  export LOCAL=false;

  # first request domain info:
  retrieveTrees

  if [ -n "$LOG_ID" ] && [ -n "$MAP_ID" ]; then
    echo "Using existing trees with MAP_ID=$MAP_ID and LOG_ID=$LOG_ID"
  fi

  COUNTER=0
  until [ -n "$LOG_ID" ] || [  $COUNTER -gt $MAX_RETRY ]; do
    # RPC was not available yet, wait and retry:
    sleep 10;
    let COUNTER+=1;
    createLog && createMap
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

function checkCmdsAvailable()
{
  if ! type jq > /dev/null 2>&1;
    then echo "Please install jq. See: https://stedolan.github.io/jq/download/"
    exit 1
  fi
}

function prepareSecrets()
{
  local EXISTS=0
  # if kt-secrets does not exist, create it:
  kubectl get secret kt-secrets
  # kubectl exits with 1 if kt-secret does not exist
  if [ $? -ne 0 ]; then
    kubectl create secret generic kt-secrets --from-file=genfiles/server.crt --from-file=genfiles/server.key 
  fi
}

# Run everything:
main
