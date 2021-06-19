#!/bin/sh

test -d data1 || mkdir -p data1
test -d data2 || mkdir -p data2
test -d data3 || mkdir -p data3

nohup ./fbox -a 10.0.0.109:8001 -b 0.0.0.0:8001 -d ./data1 -m http://10.0.0.101:8000 > fbox.log.1 2>&1 &
nohup ./fbox -a 10.0.0.109:8002 -b 0.0.0.0:8002 -d ./data2 -m http://10.0.0.101:8000 > fbox.log.2 2>&1 &
nohup ./fbox -a 10.0.0.109:8003 -b 0.0.0.0:8003 -d ./data3 -m http://10.0.0.101:8000 > fbox.log.3 2>&1 &

wait
