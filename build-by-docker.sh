#!/bin/bash

if [ "x$GOPATH" == "x" ]; then
    echo "set GOPATH first!"
    exit 1
fi

docker run -it --rm -v $GOPATH:/go -v $PWD:/work -w /work golang make $@
