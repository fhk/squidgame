#!/bin/sh
ACTUAL_DIR=$1
if diff "$ACTUAL_DIR/q1.csv" "$ACTUAL_DIR/q2.csv"; then
    echo "Parity check passed: 1+1 == 2"
    exit 0
else
    echo "Parity check failed"
    exit 1
fi
