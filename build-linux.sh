#!/bin/bash

GOOS=linux GOARCH=amd64 nodemon --ext go -x "(go build -v  -o strong-duckling-linux || exit 1)"
