#!/bin/sh
set -eu
failed=0
for lang in en-US pt-BR; do
  other=en-US; [ "$lang" = en-US ] && other=pt-BR
  for file in docs/$lang/*.md; do
    name=$(basename "$file")
    if [ ! -f "docs/$other/$name" ]; then echo "Missing docs/$other/$name"; failed=1; fi
  done
done
exit "$failed"
