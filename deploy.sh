#!/bin/bash  
# go run main.go -c -N 3 -S 5 -m 4 
tmux kill-session -t blk
name="blk"
tmux new-session -s $name -d
for i in $(seq 33)
do
    idx=`expr $i - 1`
    shardNO=`expr $idx / 3`
    nodeNO=`expr $idx % 3`
    echo "${shardNO}_${nodeNO}"
    tmux_name="$name:$i"
    tmux new-window -n "$i" -t "$name" -d
    tmux send -t $tmux_name "go run main.go -n ${nodeNO} -N 3 -s ${shardNO} -S 11 -m 4 " Enter
done