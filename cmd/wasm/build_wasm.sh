#!/usr/bin/env bash
GOOS=js GOARCH=wasm go build -o static/ws_client.wasm main.go