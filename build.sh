#!/bin/bash


OUT="./bin/"

VERSION=`date -u +%Y%m%d%H%M`
LDFLAGS="-X main.VERSION=$VERSION -s -w"
GCFLAGS=""

TIME=$(date +%s)

#ARCHS=(x64 arm7 x86 arm8 mipsle)
ARCHS=(x64)
for v in ${ARCHS[@]}; do
	go-$v build -o main.$v -ldflags "$LDFLAGS" -gcflags "$GCFLAGS" server.go

	for d in ./plugins/*/ ; do
		pushd "$d"
		go-$v build -ldflags "-pluginpath=${d//[\.\/]/-}${VERSION} -s -w" -buildmode=plugin -o plugin.so plugin.go
		popd
	done

done



