#!/bin/bash
set -e
start=`date +%s`
dir=$( dirname "$0" )

echo -e "\033[0;32mTesting\033[0m"
godep go test -cover ${dir}

echo -e "\033[0;35mTest duration : $((`date +%s`-start))s\033[0m"
