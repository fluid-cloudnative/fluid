#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$SCRIPT_DIR"

if [[ $# -lt 1 ]]; then
  echo "用法：./run_task1.sh <你的平台用户名> [输出目录]" >&2
  echo "例如：./run_task1.sh normal4" >&2
  exit 1
fi

HDFS_USER="$1"
INPUT_PATH='/user/root/Exp2/task1&2'
OUTPUT_PATH="${2:-/user/${HDFS_USER}/BG_Exp2/task1_out}"

hdfs dfs -rm -r -skipTrash "$OUTPUT_PATH" >/dev/null 2>&1 || true
hadoop jar inverted-index.jar InvertedIndex "$INPUT_PATH" "$OUTPUT_PATH"

echo "任务一运行完成。输出目录：$OUTPUT_PATH"
echo "可查看前 20 行：hdfs dfs -cat ${OUTPUT_PATH}/part-r-00000 | head -20"
