#!/bin/bash
go get github.com/golang/lint/golint
$GOPATH/bin/golint -set_exit_status $(go list ./... | grep -v '/vendor/')

