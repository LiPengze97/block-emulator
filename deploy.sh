#!/bin/bash  

tmux kill-session -t blk
name="blk"
tmux new-session -s $name -d
for i in $(seq 12)
do
    idx=`expr $i - 1`
    shardNO=`expr $idx / 3`
    nodeNO=`expr $idx % 3`
    echo "${shardNO}_${nodeNO}"
    tmux_name="$name:$i"
    tmux new-window -n "$i" -t "$name" -d
    tmux send -t $tmux_name "go run main.go -n ${nodeNO} -N 3 -s ${shardNO} -S 4 -m 3 " Enter
done