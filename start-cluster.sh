#!/bin/sh

cleanup() {
  pkill fbox
}

IP="10.0.0.101"

test -d data1 || mkdir -p data1
test -d data2 || mkdir -p data2
test -d data3 || mkdir -p data3

./fbox -a "$IP:8000" -b 0.0.0.0:8000 -d ./data1 > fbox.log.1 2>&1 &
sleep 1
./fbox -a "$IP:8001" -b 0.0.0.0:8001 -d ./data2 -m "http://$IP:8000" > fbox.log.2 2>&1 &
./fbox -a "$IP:8002" -b 0.0.0.0:8002 -d ./data3 -m "http://$IP:8000" > fbox.log.3 2>&1 &

trap cleanup EXIT

wait
