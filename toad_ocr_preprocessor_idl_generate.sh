#!/bin/bash

protoc -I rpc/idl rpc/idl/toad_ocr_preprocessor.proto --go_out=plugins=grpc:rpc/idl
sed -i '' 's/ClientConnInterface/ClientConn/g' rpc/idl/toad_ocr_preprocessor.pb.go
sed -i '' 's/SupportPackageIsVersion6/SupportPackageIsVersion4/g' rpc/idl/toad_ocr_preprocessor.pb.go
