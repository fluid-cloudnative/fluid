#!/usr/bin/env bash

# Define a timestamp function
timestamp() {
  date +"%T" # current time
}

echo Begin loading models at $(timestamp)

time cp /data/hbase-2.4.9-client-bin.tar.gz /tmp

echo Finish loading models at $(timestamp)

/helloworld