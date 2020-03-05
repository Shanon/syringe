#!/bin/bash

go get

export CGO_ENABLED=0

for GOOS in darwin linux; do
    for GOARCH in 386 amd64 arm arm64; do
        if [ "$GOOS" != "linux" ]; then
            if echo "$GOARCH" | grep "arm" > /dev/null; then
                break
            fi
        fi
        export GOOS
        export GOARCH
        go build -v -o output/syringe_${GOOS}_${GOARCH} main.go
    done
done

chown $UIDGID output/*
