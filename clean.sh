#!/bin/bash
set -e
dir=$( dirname "$0" )

echo -e "\033[0;32mCleaning\033[0m"
rm -Rf ${dir}/dist