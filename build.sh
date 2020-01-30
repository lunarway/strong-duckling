#!/bin/bash

nodemon --ext go -x "(go build -v -o strong-duckling || exit 1)"
