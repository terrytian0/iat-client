#!/bin/bash
docker run -idt -e "IAT_SERVER=192.168.3.10:8080" -e "IAT_CLIENT_NAME=192.168.3.10" terrytian/iat-client:v0.2
