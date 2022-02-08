#!/bin/sh

# requirements
#   go get -u google.golang.org/grpc
#   go get -u google.golang.org/protobuf/proto
#
#   sudo dnf install protoc
#
#   go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway
#   go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger

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
    --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative,require_unimplemented_servers=false \
    "${file}"

  if [ $? -ne 0 ]
  then
    echo "${0}: fail attempting to generate the protos for file ${file}"
    exit 1
  fi

  # Generate the REST reverse proxy.
  protoc -I/usr/local/include -I. \
    --grpc-gateway_out=logtostderr=true,paths=source_relative,grpc_api_configuration=rest-annotations.yaml:. \
    "${file}"

  # Finally, generate the swagger file which describes the REST API in detail.
  protoc -I/usr/local/include -I. \
    --swagger_out=logtostderr=true,grpc_api_configuration=rest-annotations.yaml:. \
    "${file}"
done
