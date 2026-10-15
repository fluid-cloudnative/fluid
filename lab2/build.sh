#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$SCRIPT_DIR"

if command -v hadoop >/dev/null 2>&1; then
  HADOOP_CMD="$(command -v hadoop)"
elif [[ -n "${HADOOP_HOME:-}" && -x "$HADOOP_HOME/bin/hadoop" ]]; then
  HADOOP_CMD="$HADOOP_HOME/bin/hadoop"
else
  echo "错误：未找到 hadoop 命令。请先把 hadoop 加到 PATH，或设置 HADOOP_HOME。" >&2
  exit 1
fi

HADOOP_CP="$($HADOOP_CMD classpath)"
rm -rf build inverted-index.jar
mkdir -p build/classes

javac -encoding UTF-8 -cp "$HADOOP_CP" -d build/classes InvertedIndex.java
jar cfe inverted-index.jar InvertedIndex -C build/classes .

echo "构建完成：$SCRIPT_DIR/inverted-index.jar"
