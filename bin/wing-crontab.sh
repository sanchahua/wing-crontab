#!/usr/bin/env bash
current_path=$(cd `dirname $0`; pwd)
echo "$$" > ${current_path}/wing-crontab-sh.pid
while [ 1 ]; do
    ${current_path}/wing-crontab
done