#!/usr/bin/env bash
MAX_JOBS=30

for i in `seq 1 $MAX_JOBS`; do
    curl http://localhost:6666 &
done

wait
echo all processes complete
