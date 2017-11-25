#!/bin/sh
set -e

PROJECT_DIR="$(cd "$(dirname "$0")/.."; pwd)"
LINUX_FILENAME="app-metrics-plugin-linux"
MAC_FILENAME="app-metrics-plugin-darwin"
WIN_FILENAME="app-metrics-plugin.exe"

if [[ "$1" = "release" ]] ; then
	TAG="$2"
	: ${TAG:?"Usage: build_all.sh [release] [TAG]"}


	if git tag | grep $TAG > /dev/null 2>&1 ; then
		echo "$TAG exists, remove it or increment"
		exit 1
	else
		MAJOR=`echo $TAG | sed 's/^v//' | awk 'BEGIN {FS = "." } ; { printf $1;}'`
		MINOR=`echo $TAG | sed 's/^v//' | awk 'BEGIN {FS = "." } ; { printf $2;}'`
		BUILD=`echo $TAG | sed 's/^v//' | awk 'BEGIN {FS = "." } ; { printf $3;}'`

		`sed -i "" -e "1,/Major:.*/s/Major:.*/Major: $MAJOR,/" \
			-e "1,/Minor:.*/s/Minor:.*/Minor: $MINOR,/" \
			-e "1,/Build:.*/s/Build:.*/Build: $BUILD,/" ${PROJECT_DIR}/cmd/plugin/app_metrics.go`
	fi
fi


GOOS=linux GOARCH=amd64 go build -o $LINUX_FILENAME ${PROJECT_DIR}/cmd/plugin/app_metrics.go
LINUX64_SHA1=`cat $LINUX_FILENAME | openssl sha1`
mkdir -p bin/linux64
mv $LINUX_FILENAME bin/linux64

GOOS=darwin GOARCH=amd64 go build -o $MAC_FILENAME ${PROJECT_DIR}/cmd/plugin/app_metrics.go
OSX_SHA1=`cat $MAC_FILENAME | openssl sha1`
mkdir -p bin/osx
mv $MAC_FILENAME bin/osx

GOOS=windows GOARCH=amd64 go build -o $WIN_FILENAME ${PROJECT_DIR}/cmd/plugin/app_metrics.go
WIN64_SHA1=`cat $WIN_FILENAME | openssl sha1`
mkdir -p bin/win64
mv $WIN_FILENAME bin/win64

NOW=`TZ=UC date +'%Y-%m-%dT%TZ'`

cat repo-index.yml |
sed "s/__osx-sha1__/$OSX_SHA1/" |
sed "s/__win64-sha1__/$WIN64_SHA1/" |
sed "s/__linux64-sha1__/$LINUX64_SHA1/" |
sed "s/__TAG__/$TAG/" |
sed "s/__TODAY__/$NOW/" |
cat

if [[ "$1" = "release" ]] ; then
	git commit -am "Build version $TAG"
	git tag -a $TAG -m "AppMetrics Plugin v$TAG"
	echo "Tagged release, 'git push --follow-tags' to push it to github, upload the binaries to github"
	echo "and copy the output above to the cli repo you plan to deploy in"
fi
