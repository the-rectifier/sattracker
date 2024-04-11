#!/usr/bin/bash

docker build -t n2yo-proxy .
docker run -d -p 9443:9443 n2yo-proxy
