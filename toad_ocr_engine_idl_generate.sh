#!/bin/bash

protoc -I client/toad_ocr_engine/idl client/toad_ocr_engine/idl/toad_ocr.proto --go_out=plugins=grpc:client/toad_ocr_engine/idl
sed -i '' 's/ClientConnInterface/ClientConn/g' client/toad_ocr_engine/idl/toad_ocr.pb.go
sed -i '' 's/SupportPackageIsVersion6/SupportPackageIsVersion4/g' client/toad_ocr_engine/idl/toad_ocr.pb.go
