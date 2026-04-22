#!/bin/bash

WORKDIR=/go/src/github.com/ooclab/otunnel
docker run -it --rm -v $PWD:$WORKDIR -w $WORKDIR golang:1.26 make $@
