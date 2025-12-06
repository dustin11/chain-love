#!/bin/bash

ENV="uat"
SERVICE="119.45.241.154"

BASEDIR=$(dirname $(cd `dirname $0`; pwd))
echo "basedir $BASEDIR"
sh $BASEDIR/bin/release.sh $BASEDIR $SERVICE $ENV