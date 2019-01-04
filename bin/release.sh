#!/usr/bin/env bash
cp -rf ./config_online/* ./config/
tar -czvf wing-crontab.tar.gz ./*
cp -rf ./config_dev/* ./config/