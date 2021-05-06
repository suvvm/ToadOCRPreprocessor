#!/bin/bash

RUN_NAME="toad_ocr_preprocessor"

if [ -n "$GOPATH" ];then
   rm -rf ./output/*  # 清空out目录
   rm -f ${RUN_NAME}
   mkdir -p ./output/bin # 创建二进制文件存放目录
   mkdir -p ./output/images
   go build -o ./output/bin/${RUN_NAME}
   cp ./output/bin/${RUN_NAME} ${RUN_NAME}
else
	echo "GOPATH is needed!"
fi