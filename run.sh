nohup python3 rf.py --port 5009 > rfout.txt &
nohup ./findServer -rf 5009 -filter $1 > findout.txt &