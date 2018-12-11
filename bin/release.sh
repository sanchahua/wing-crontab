#!/usr/bin/env bash
cp ./config_online/* ./config/
tar -czvf xcrontab.tar.gz ./*
cp ./config_dev/* ./config/