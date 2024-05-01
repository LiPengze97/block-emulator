#!/bin/bash  
# go run main.go -c -N 3 -S 5 -m 3 
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