#!/bin/bash

NAME=jard

go build -o $NAME -ldflags "-linkmode external -extldflags -static"

# Grab the version, per the build
VERSION=`./${NAME} --version | perl -pe 's/\n/\t/g' | cut -f 1 | cut --delimiter=' ' -f 2`

echo
echo "$NAME $VERSION..."

rm -f $NAME.zip
rm -Rf builds/
mkdir builds
cp $NAME builds/
zip -j $NAME.zip builds/*

rm -Rf builds/
