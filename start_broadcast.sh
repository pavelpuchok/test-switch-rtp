#!/usr/bin/env bash

#ffmpeg -re -stream_loop -1 -i "$1" -map 0:v -vcodec libvpx -g 30 -f rtp "rtp://127.0.0.1:$2?pkt_size=1200"
ffmpeg \
    -re \
    -stream_loop -1 \
    -i "$1" \
    -map 0:v \
    -c copy \
    -f rtp "rtp://127.0.0.1:$2?pkt_size=1200"
