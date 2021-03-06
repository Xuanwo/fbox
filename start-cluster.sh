#!/bin/sh

cleanup() {
  pkill fbox
}

if [ -z "$IP" ]; then
  IP="127.0.0.1"
fi

test -d data1 || mkdir -p data1
test -d data2 || mkdir -p data2
test -d data3 || mkdir -p data3

if [ x"$DEBUG" = x"1" ]; then
  printf "Running fbox master in debug mode...\n"
  ./fbox -D -a "$IP:8000" -b 0.0.0.0:8000 -d ./data1 > fbox.log.1 2>&1 &
else
  ./fbox -a "$IP:8000" -b 0.0.0.0:8000 -d ./data1 > fbox.log.1 2>&1 &
fi

sleep 1
./fbox -a "$IP:8001" -b 0.0.0.0:8001 -d ./data2 -m "http://$IP:8000" > fbox.log.2 2>&1 &
./fbox -a "$IP:8002" -b 0.0.0.0:8002 -d ./data3 -m "http://$IP:8000" > fbox.log.3 2>&1 &

trap cleanup EXIT

wait
