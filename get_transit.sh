#!/bin/sh

wget https://rrgtfsfeeds.s3.amazonaws.com/gtfs_subway.zip &&
rm -rf gtfs_subway/*.txt &&
unzip -d gtfs_subway gtfs_subway.zip &&
rm gtfs_subway.zip
