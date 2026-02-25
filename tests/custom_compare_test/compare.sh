#!/bin/sh
ACTUAL_DIR=$1
EXPECTED_DIR=$2

echo "Checking $ACTUAL_DIR against $EXPECTED_DIR"

if [ ! -f "$ACTUAL_DIR/output.txt" ]; then
    echo "output.txt not found in actual results"
    exit 1
fi

if grep -q "custom content" "$ACTUAL_DIR/output.txt"; then
    echo "Content matched"
    exit 0
else
    echo "Content did not match"
    exit 1
fi
