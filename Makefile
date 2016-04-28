
#
# Copyright 2010 The Go Authors.  All rights reserved.
# https://github.com/golang/protobuf
#
# Redistribution and use in source and binary forms, with or without
# modification, are permitted provided that the following conditions are
# met:
#
#     * Redistributions of source code must retain the above copyright
# notice, this list of conditions and the following disclaimer.
#     * Redistributions in binary form must reproduce the above
# copyright notice, this list of conditions and the following disclaimer
# in the documentation and/or other materials provided with the
# distribution.
#     * Neither the name of Google Inc. nor the names of its
# contributors may be used to endorse or promote products derived from
# this software without specific prior written permission.
#
# THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
# "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
# LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
# A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
# OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
# SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
# LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
# DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
# THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
# (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
# OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.


PROTOINCLUDE ?= /usr/local/include

#include $(GOHOME)/src/pkg/github.com/golang/protobuf/Make.protobuf
DEPS:= $(shell find . -type f -name '*.proto' | sed 's/proto$$/pb.go/')
OUTPUT:= $(GOPATH)/src
FLAGS+= --go_out=plugins=grpc
INCLUDES+= -I=.
INCLUDES+= -I=$(GOPATH)/src/
INCLUDES+= -I=$(PROTOINCLUDE)


main: proto
	go build ./
	go build -o e2e-key-signer ./backend

# The list of returned packages might not be unique. Fortunately go test gets
# rid of duplicate.
test: main
	go test `find . | grep '_test\.go$$' | sort | xargs -n 1 dirname`
	python tests/api_proxy_test.py

fmt:
	gofmt -w `find . | grep -e '\.go$$'`

proto: $(DEPS)

./%.pb.go:  %.proto
	protoc $(INCLUDES) $(FLAGS),:. $(dir $<)*.proto

clean:
	rm -f $(DEPS)
	rm -f srv e2e-key-server e2e-key-signer
	rm -rf infra*
	rm -f tree-db.sqlite3

