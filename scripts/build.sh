#!/bin/bash
set -e

GITVERSION=`git describe --always`
SOURCEDIR=$1
BUILDDIR=$SOURCEDIR/build
PLUGIN=`echo $SOURCEDIR | grep -oh "snap-.*"`
ROOTFS=$BUILDDIR/rootfs
BUILDCMD='go build -a -ldflags "-w"'

echo
echo "****  Snap Plugin Build  ****"
echo
echo "Source dir: $SOURCEDIR"
echo "Building snap plugin: $PLUGIN"

export CGO_ENABLED=0
rm -rf $ROOTFS/*
mkdir -p $ROOTFS
$BUILDCMD -o $ROOTFS/$PLUGIN
