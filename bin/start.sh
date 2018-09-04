#!/usr/bin/env bash
current_path=$(cd `dirname $0`; pwd)
chmod 0777 ${current_path}/xcrontab
chmod 0777 ${current_path}/xcrontab.sh
nohup ${current_path}/xcrontab.sh >debug.log &