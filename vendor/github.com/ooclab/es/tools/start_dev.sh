#!/bin/bash

DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
ES="${DIR}/../"
docker run -it --rm -v $DIR/.bashrc:/root/.bashrc -v /data/tmp/es/:/go/src/ -v $ES:/go/src/github.com/ooclab/es golang bash
