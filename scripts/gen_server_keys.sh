#!/bin/bash

# Copyright 2016 Google Inc. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

COMMONNAME=""
ADDRESS=""
SUBJECT="/C=US"

while getopts d:a: option; do
    case "${option}" in
	d) COMMONNAME=${OPTARG};;
	a) ADDRESS=${OPTARG};;
	*) echo "usage: ./generate.sh -d <domain> -a <ip_address>"; exit 1;;
    esac
done

if [[ -n "${COMMONNAME}" ]]; then
    SUBJECT="${SUBJECT}/CN=${COMMONNAME}"
fi

SANEXT="[SAN]\nbasicConstraints=CA:TRUE\nsubjectAltName=DNS:localhost"
if [[ -n "${ADDRESS}" ]]; then
    SANEXT="${SANEXT},IP.2:${ADDRESS}"
fi

# Create output directory.
mkdir -p "${GOPATH}/src/github.com/google/keytransparency/genfiles"
cd "${GOPATH}/src/github.com/google/keytransparency/genfiles"

# Generate TLS keys.
openssl genrsa -des3 -passout pass:x -out server.pass.key 2048
openssl rsa -passin pass:x -in server.pass.key -out server.key
chmod 600 server.key
rm server.pass.key
openssl req -new \
	-key server.key \
	-subj "${SUBJECT}" \
	-reqexts SAN \
	-config <(cat /etc/ssl/openssl.cnf \
		      <(printf "${SANEXT}")) \
	-out server.csr
openssl x509 -req -days 365 -in server.csr -signkey server.key \
	-out server.crt -extensions SAN \
	-extfile <(printf "${SANEXT}")

# Generate VRF keys.
openssl ecparam -name prime256v1 -genkey -noout -out vrf-key.pem
chmod 600 vrf-key.pem
openssl ec -in vrf-key.pem -pubout -out vrf-pubkey.pem

