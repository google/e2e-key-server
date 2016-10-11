# Key Transparency
Key Transparency is a distributed implementation of
[CONIKS](https://eprint.iacr.org/2014/1004.pdf), written in Go.

## Running a Key Transparency Cluster
1. Install prerequisites

  ```sh
  go get -u github.com/mattn/goreman
  go get -u github.com/coreos/etcd
  go get -u github.com/google/key-transparency/cmd/...
  ```
  Ensure `$GOBIN` is in your `$PATH:`

2. Generate test keys

  ```sh
  make -C testdata
  ```
3. Get [Application Default Credentials](https://developers.google.com/identity/protocols/application-default-credentials)

  Use this [wizard](https://console.developers.google.com/start/api?id=e2ekeys)
  and set ```GOOGLE_APPLICATION_CREDENTIALS``` environment variable in
  [.env](.env) to point to the credentials file.
4. Download and run an XJSON [Certificate Transparency](https://github.com/google/certificate-transparency) Server.

  Set the `CTLOG` variable to the URL of the CT server URL in the [.env](.env) file.

5. Run

  ```sh
  goreman start
  ```

## Using the Key Transparency Client
1. Get a client secret
  Use this [wizard](https://console.developers.google.com/start/api?id=e2ekeys)
  and set the `client_secret` path in `.key-transparency.yaml`

2. Set / Update a user's keys:

  ```
  ./key-transparency-client post <email> -d '{"app1": "dGVzdA=="}' --config=./.key-transparency.yaml
  {Keys:map[app1:[116 101 115 116]}
  ```

2. Fetch and verify a user's keys:

  ```
  ./key-transparency-client get <email> --config=.key-transparency.yaml --verbose
  ✓ Commitment verified.
  ✓ VRF verified.
  ✓ Sparse tree proof verified.
  ✓ Signed Map Head signature verified.
  CT ✓ STH signature verified.
  CT ✓ Consistency proof verified.
  CT   New trusted STH: 2016-09-12 15:31:19.547 -0700 PDT
  CT ✓ SCT signature verified. Saving SCT for future inclusion proof verification.
  ✓ Signed Map Head CT inclusion proof verified.
  keys:<key:"app1" value:"test" >
  ```

  ```
  ./key-transparency-client history <email> --config=.key-transparency.yaml
  Epoch |Timestamp                    |Profile
  4     |Mon Sep 12 22:23:54 UTC 2016 |keys:<key:"app1" value:"test" >
  ```

## Building from scratch
1. Install [ProtocolBuffers](https://github.com/golang/protobuf#installation) 3.0 or later.
  ```sh
  mkdir tmp
  cd tmp
  git clone https://github.com/google/protobuf
  cd protobuf
  ./autogen.sh
  ./configure
  make
  make check
  sudo make install
  ```

2. Then, ```go get -u``` as usual

  ```sh
  go get -u github.com/google/key-transparency/cmd/...
  go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway
  go get -u github.com/golang/protobuf/protoc-gen-go
  ```

## Projects Using Key Transparency
* [Google End-To-End](https://github.com/google/end-to-end).
