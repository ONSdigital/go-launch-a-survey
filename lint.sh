#!/bin/bash
go get -u golang.org/x/lint/golint
$GOPATH/bin/golint -set_exit_status $(go list ./... | grep -v '/vendor/')

