#!/bin/sh

PORT=8080
LATENCY=2000ms
BANDWIDTH=0.1mbit
LOSS=20%

start() {
    tc qdisc add dev lo root handle 1: prio
    
    tc filter add dev lo protocol ip parent 1:0 prio 1 u32 \
        match ip dport $PORT 0xffff flowid 1:1
    
    tc qdisc add dev lo parent 1:1 handle 10: netem \
        delay $LATENCY loss $LOSS rate $BANDWIDTH
    
    echo "Network conditions applied to port $PORT"
    echo "Latency: $LATENCY, Bandwidth: $BANDWIDTH, Loss: $LOSS"
}

stop() {
    tc qdisc del dev lo root 2>/dev/null
    echo "Network conditions removed"
}

case "$1" in
    start) start ;;
    stop) stop ;;
    *) echo "Usage: $0 {start|stop}" ;;
esac
