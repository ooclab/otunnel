#!/bin/bash


BASENAME=otunnel
VERSION=1.0.1
GOBUILD="go build"


unalias -a


#for GOOS in darwin linux windows; do
#    for GOARCH in amd64 386; do
#        PROGRAM_NAME=${BASENAME}-${GOOS}-${GOARCH}-${VERSION}
#
#        [ -f $PROGRAM_NAME ] && rm $PROGRAM_NAME
#        [ -f ${PROGRAM_NAME}.gz ] && rm ${PROGRAM_NAME}.gz
#
#        GOOS=$GOOS GOARCH=$GOARCH $GOBUILD -v -ldflags "-s -X main.buildstamp=`date '+%Y-%m-%d_%H:%M:%S_%z'` -X main.githash=`git rev-parse HEAD`" -o $PROGRAM_NAME
#        gzip $PROGRAM_NAME
#    done
#done

for TARGET in mac linux-64 linux-32 windows-64 windows-32; do
    ./build-by-docker.sh $TARGET
    DN=${BASENAME}-${VERSION}-${TARGET}
    mkdir -pv $DN
    if [ "${TARGET#windows}" != "$TARGET" ]; then
        mv otunnel $DN/otunnel.exe
    else
        mv otunnel $DN
    fi
    cp docs/USAGE.md $DN
    cp docs/CHANGELOG.md $DN
    tar cvzf $DN.tar.gz $DN
    rm -rvf $DN
done
