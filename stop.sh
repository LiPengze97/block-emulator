#!/bin/bash  
name="blk"
for i in $(seq 12)
do
    tmux_name="$name:$i"
    tmux send -t $tmux_name C-c
done