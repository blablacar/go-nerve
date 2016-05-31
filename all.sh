#!/bin/bash
set -e
start=`date +%s`
dir=$( dirname "$0" )

${dir}/clean.sh
${dir}/build.sh
${dir}/quality.sh
PATH=${GOPATH}/bin:$PATH ${dir}/test.sh

echo -e "\033[0;35mAll duration : $((`date +%s`-start))s\033[0m"
