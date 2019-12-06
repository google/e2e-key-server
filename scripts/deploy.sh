#!/usr/bin/env bash
#set -o pipefail
#set -o errexit
#set -o nounset
#set -o xtrace
set -eufx

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


export PROJECT_NAME_CI=key-transparency
export CLOUDSDK_COMPUTE_ZONE=us-central1-a
export CLUSTER_NAME_CI=ci-cluster


gcloud --quiet config set project ${PROJECT_NAME_CI}
gcloud --quiet config set compute/zone ${CLOUDSDK_COMPUTE_ZONE}
gcloud --quiet config set container/cluster ${CLUSTER_NAME_CI}
gcloud --quiet container clusters get-credentials ${CLUSTER_NAME_CI}
gcloud --quiet auth configure-docker

# Test current directory before deleting anything
test $(basename $(pwd)) == "keytransparency" || exit 1

# kubectl exits with 1 if kt-secret does not exist
if ! kubectl get secret kt-secrets; then
  echo "Generating keys..."
  rm -f ./genfiles/*
  ./scripts/prepare_server.sh -f
  kubectl create secret generic kt-secrets --from-file=genfiles/server.crt --from-file=genfiles/server.key --from-file=genfiles/monitor_sign-key.pem
fi

echo "Building docker images..."
docker-compose build


echo "Pushing docker images..."
docker-compose push

echo "Tagging docker images..."
gcloud --quiet container images add-tag gcr.io/${PROJECT_NAME_CI}/prometheus:${TRAVIS_COMMIT} gcr.io/${PROJECT_NAME_CI}/prometheus:latest
gcloud --quiet container images add-tag gcr.io/${PROJECT_NAME_CI}/keytransparency-server:${TRAVIS_COMMIT} gcr.io/${PROJECT_NAME_CI}/keytransparency-server:latest
gcloud --quiet container images add-tag gcr.io/${PROJECT_NAME_CI}/keytransparency-sequencer:${TRAVIS_COMMIT} gcr.io/${PROJECT_NAME_CI}/keytransparency-sequencer:latest
gcloud --quiet container images add-tag gcr.io/${PROJECT_NAME_CI}/keytransparency-monitor:${TRAVIS_COMMIT} gcr.io/${PROJECT_NAME_CI}/keytransparency-monitor:latest


echo "Updating jobs..."
kubectl apply -k deploy/kubernetes/overlays/gke
kubectl set image deploy/prometheus prometheus=gcr.io/${PROJECT_NAME_CI}/prometheus:${TRAVIS_COMMIT}
kubectl set image deploy/log-server log-server=gcr.io/${PROJECT_NAME_CI}/log-server:${TRAVIS_COMMIT}
kubectl set image deploy/log-signer log-signer=gcr.io/${PROJECT_NAME_CI}/log-signer:${TRAVIS_COMMIT}
kubectl set image deploy/map-server map-server=gcr.io/${PROJECT_NAME_CI}/map-server:${TRAVIS_COMMIT}
kubectl set image deploy/server server=gcr.io/${PROJECT_NAME_CI}/keytransparency-server:${TRAVIS_COMMIT}
kubectl set image deploy/sequencer sequencer=gcr.io/${PROJECT_NAME_CI}/keytransparency-sequencer:${TRAVIS_COMMIT}
kubectl set image deploy/monitor monitor=gcr.io/${PROJECT_NAME_CI}/keytransparency-monitor:${TRAVIS_COMMIT}
