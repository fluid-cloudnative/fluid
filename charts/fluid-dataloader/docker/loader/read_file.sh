#! /usr/bin/env bash

echo THREADS=$THREADS
echo DATAPATH=$DATA_PATH
echo python multithread_read_benchmark.py --threads=$THREADS --path=$DATA_PATH
python multithread_read_benchmark.py --threads=$THREADS --path=$DATA_PATH
