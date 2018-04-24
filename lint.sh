#!/bin/bash
mkdir -p $GOPATH/src/golang.org/x
git clone --dept=1 https://github.com/golang/lint.git $GOPATH/src/golang.org/x/lint
go get github.com/golang/lint/golint
$GOPATH/bin/golint -set_exit_status $(go list ./... | grep -v '/vendor/')

