#!/bin/bash

VERSION=$(git describe --always --long)
go get
GOOS=linux GOARCH=386 CGO_ENABLED=0 go build -i -v -ldflags="-X main.version=${VERSION}"
docker build -t stevelacy/kubermaster .
