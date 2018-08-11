#!/usr/bin/env bash
current_path=$(cd `dirname $0`; pwd)
chmod 0777 ${current_path}/wing-crontab
chmod 0777 ${current_path}/wing-crontab.sh
nohup ${current_path}/wing-crontab.sh >/dev/null &