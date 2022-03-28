#!/usr/bin/env bash

# Define a timestamp function
timestamp() {
  date +"%T" # current time
}

echo Begin loading models at $(timestamp)

time cp -r /data/ /tmp

echo Finish loading models at $(timestamp)

/helloworld
