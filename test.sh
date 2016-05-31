#!/bin/bash
set -e
dir=$( dirname "$0" )

echo -e "\033[0;32mTesting\033[0m"
godep go test -cover ${dir}
