# Simple ping utility written in Go
Build:
```
go build
```
Usage:
```
pinger [Options] ya.ru
```
# Supported features
Makes ICMP IPv4 echo requests in a loop with given interval. For each echo reply utility reports time passed between request and reply. At the end of execution it reports packet loss and echo statistics.
* `-W <duration>` limits response wait time
* `-c <count>` limits number of sent requests
* `-i <interval>` specifies interval between request the pinger sends
* `-t <count>` Option to set Time to leave
