#!/bin/bash  
name="blk"
for i in $(seq 132)
do
    tmux_name="$name:$i"
    tmux send -t $tmux_name C-c
done