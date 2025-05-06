#!/bin/bash

# Duration in seconds (20 minutes = 1200 seconds)
duration=1200
start_time=$(date +%s)
end_time=$((start_time + duration))

while [ $(date +%s) -lt $end_time ]
do
    # Generate random number of concurrent requests between 3 and 50
    concurrent_requests=$(( $RANDOM % 48 + 3 ))
    
    # Launch concurrent requests
    for ((i=1; i<=concurrent_requests; i++))
    do
        curl -s "http://localhost:8080/rolldice/Alice" > /dev/null &
    done
    
    # Random sleep between 100ms and 1s before next batch
    sleep 0.$(( $RANDOM % 100 + 100 ))
    
    # Wait for all background curl processes to complete
    wait
done
