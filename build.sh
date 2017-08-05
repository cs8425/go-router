#!/bin/bash


OUT="./bin/"

VERSION=`date -u +%Y%m%d%H%M`
LDFLAGS="-X main.VERSION=$VERSION -s -w"
GCFLAGS=""

TIME=$(date +%s)

GO () {
	TYPE=$1
	shift

	case "$TYPE" in
	'x64')
		GOOS=linux GOARCH=amd64 go "$@"
	;;
	'x86')
		GOOS=linux GOARCH=386 go "$@"
	;;
	'arm5')
		GOOS=linux GOARCH=arm GOARM=5 go "$@"
	;;
	'arm7')
		GOOS=linux GOARCH=arm GOARM=7 go "$@"
	;;
	'arm8')
		GOOS=linux GOARCH=arm64 go "$@"
	;;
	'win32')
		GOOS=windows GOARCH=386 go "$@"
	;;
	'win64')
		GOOS=windows GOARCH=amd64 go "$@"
	;;
	'mac64')
		GOOS=darwin GOARCH=amd64 go "$@"
	;;
	'mipsle')
		GOOS=linux GOARCH=mipsle go "$@"
	;;
	'mips')
		GOOS=linux GOARCH=mips go "$@"
	;;

	esac
}

#ARCHS=(x64 arm7 x86 arm8 mipsle)
ARCHS=(x64)
for v in ${ARCHS[@]}; do

#	go-$v build -o main.$v -ldflags "$LDFLAGS" -gcflags "$GCFLAGS" *.go
	GO $v build -o main.$v -ldflags "$LDFLAGS" -gcflags "$GCFLAGS" *.go

	for d in ./plugins/*/ ; do
#		break
		pushd "$d"
#		go-$v build -ldflags "-pluginpath=${d//[\.\/]/-}${VERSION} -s -w" -buildmode=plugin -o plugin.so plugin.go
		GO $v build -ldflags "-pluginpath=${d//[\.\/]/-}${VERSION} -s -w" -buildmode=plugin -o plugin.so plugin.go
		popd
	done

done



