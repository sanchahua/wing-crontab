#!/usr/bin/env bash
cp -rf ./config_online/* ./config/
tar -czvf xcrontab.tar.gz ./*
cp -rf ./config_dev/* ./config/