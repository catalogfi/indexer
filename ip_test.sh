#!/bin/bash

# Array of IP addresses
ip_addresses=(
    '78.30.33.83'
    '69.197.185.106'
    '65.0.196.2'
    # ... rest of the IP addresses
)

# Function to measure latency
measure_latency() {
    ip_address=\$1
    latency=$(ping -c 4 -q $ip_address | awk -F'/' '/^rtt/ {print \$5}')
    echo $latency
}

# Measure latencies and store them in an associative array
declare -A latencies
for ip_address in "${ip_addresses[@]}"; do
    latency=$(measure_latency $ip_address)
    latencies[$ip_address]=$latency
done

# Sort IP addresses based on latency
sorted_ip_addresses=($(
    for key in "${!latencies[@]}"; do
        echo "$key ${latencies[$key]}"
    done | sort -n -k2 | awk '{print \$1}'
))

# Select the IP address with the lowest latency
fastest_ip=${sorted_ip_addresses[0]}

echo "Fastest IP address: $fastest_ip"
