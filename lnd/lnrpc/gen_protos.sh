#!/bin/sh

# requirements
#   go get -u google.golang.org/grpc
#   go get -u google.golang.org/protobuf/proto
#
#   sudo dnf install protoc
#
#   go install github.com/pseudomuto/protoc-gen-doc/cmd/protoc-gen-doc@latest

echo "Generating root gRPC server protos"

PROTOS="rpc.proto walletunlocker.proto metaservice.proto pkt.proto **/*.proto"

# For each of the sub-servers, we then generate their protos, but a restricted
# set as they don't yet require REST proxies, or swagger docs.
for file in $PROTOS; do
  DIRECTORY=$(dirname "${file}")

  if [ "${DIRECTORY}" == '.' ]
  then
    DIRECTORY="$( pwd )"
  else
    DIRECTORY="$( pwd )/${DIRECTORY}"
  fi

  echo "Generating protos from ${file}, into ${DIRECTORY}"

  # Generate the protos.
  protoc -I/usr/local/include -I. \
    -I="${GOPATH}/pkg/mod/github.com/gogo/protobuf@v1.3.2/protobuf" \
    --go_out=. \
    --go_opt=paths=source_relative \
    --go-grpc_out=. \
    --go-grpc_opt=paths=source_relative,require_unimplemented_servers=false \
    --doc_out=. \
    --doc_opt="json,${file}.doc.json" \
    "${file}"

  if [ $? -ne 0 ]
  then
    echo "${0}: fail attempting to generate the protos for file ${file}"
    exit 1
  fi

done

go run ../pkthelp/mkhelp/mkhelp.go . > ../pkthelp/pkthelp_gen.go
if [ $? -ne 0 ]
then
  echo "${0}: fail attempting to generate help for protos"
  exit 1
fi