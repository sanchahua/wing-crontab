#!/usr/bin/env bash
current_path=$(cd `dirname $0`; pwd)
kill -9 `cat ./wing-crontab.pid`