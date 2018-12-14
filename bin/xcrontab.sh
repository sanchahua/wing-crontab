#!/usr/bin/env bash
current_path=$(cd `dirname $0`; pwd)
echo "$$" > ${current_path}/xcrontab-sh.pid
while [ 1 ]; do
    ${current_path}/xcrontab -l 0.0.0.0:48019
done