#!/bin/bash

nodemon --signal SIGTERM --watch /strong-duckling -x "/strong-duckling $@ || exit 1"
