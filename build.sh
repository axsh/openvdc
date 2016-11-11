#!/bin/bash

set -e

go build -v ./cmd/openvdc
go build -v ./cmd/openvdc-scheduler
go build -v ./cmd/openvdc-executor
