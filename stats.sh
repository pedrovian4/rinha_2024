#!/bin/bash

services=$(docker-compose ps -q)

while true; do
    clear
    docker stats --no-stream $services | awk '{if (NR==1) print "\033[1;36m" $0 "\033[0m"; else print "\033[1;34m" $0 "\033[0m"}'
    sleep 10
done