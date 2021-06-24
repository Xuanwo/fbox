#!/bin/sh

cleanup() {
  pkill fbox
}

test -d data1 || mkdir -p data1
test -d data2 || mkdir -p data2
test -d data3 || mkdir -p data3

./fbox -a 10.0.0.104:8000 -b 0.0.0.0:8000 -d ./data1 > fbox.log.1 2>&1 &
./fbox -a 10.0.0.104:8001 -b 0.0.0.0:8001 -d ./data2 -m http://10.0.0.104:8000 > fbox.log.2 2>&1 &
./fbox -a 10.0.0.104:8002 -b 0.0.0.0:8002 -d ./data3 -m http://10.0.0.104:8000 > fbox.log.3 2>&1 &

trap cleanup EXIT

wait
