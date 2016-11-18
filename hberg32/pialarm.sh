#!/bin/bash
kill $(ps aux | grep '[p]ialarm.py' | awk '{print $2}')
logdir=/var/log/pialarm
pialarmdir=/home/pi/pialarm
datetime=`date +%m%d%Y%H%M`
mv $logdir/pialarm.log $logdir/pialarm.$datetime
amixer set PCM "90%"
python -u $pialarmdir/pialarm.py >> $logdir/pialarm.log 2>&1
