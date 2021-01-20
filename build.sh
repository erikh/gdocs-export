#!/bin/sh

rm -rf release

for GOOS in darwin linux
do
  for GOARCH in 386 amd64
  do
    export GOOS GOARCH
    go build -v -o release/gdexport-$VERSION-$GOOS-$GOARCH ./cmd/gdexport 
    gzip -9 release/gdexport-$VERSION-$GOOS-$GOARCH
  done
done
