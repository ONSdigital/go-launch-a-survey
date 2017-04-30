#!/bin/bash
go fmt $(go list ./... | grep -v '/vendor/')