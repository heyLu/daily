#!/bin/bash

ARGS="${ARGS:-}"
RUN_DIR="${RUN_DIR:-/srv/daily}"
BINARY_PATH="${BINARY_PATH:-${RUN_DIR}/daily}"
RUN_USER="${RUN_USER:-daily}"
RUN_GROUP="${RUN_GROUP:-daily}"

cat <<-EOF
[Unit]
Description=daily - track data about points in time

[Service]
User=$RUN_USER
Group=$RUN_GROUP
ExecStart=$BINARY_PATH $ARGS
ReadWritePaths=${RUN_DIR}
WorkingDirectory=$RUN_DIR

[Install]
WantedBy=multi-user.target
EOF
