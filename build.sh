#!/bin/bash

nodemon --ext go -x "go build -v || exit 1"
