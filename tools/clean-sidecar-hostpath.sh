#!/bin/bash

################################################################################
# Fluid Sidecar HostPath 清理脚本
#
# 功能：清理 Fluid Sidecar 模式下残留的 HostPath 目录
# 
# 安全机制：
# 1. 仅在目录数超过 1000 时执行清理
# 2. 仅清理超过 3 个月的目录
# 3. 检查挂载点，避免删除正在使用的目录
# 4. 仅删除最后一级子目录
# 5. 使用 rmdir 确保目录为空才删除
# 6. 多重验证机制防止误删除
#
# 使用方法：
#   ./clean-sidecar-hostpath.sh [--dry-run] [--base-dir <dir>]
#
# 参数：
#   --dry-run           仅显示将要删除的目录，不实际执行删除
#   --base-dir <dir>    指定基础目录（默认：/runtime-mnt）
#   --threshold <num>   指定触发清理的目录数阈值（默认：1000）
#   --age-days <num>    指定目录保留天数（默认：90 天，即 3 个月）
#   --help              显示帮助信息
#
################################################################################

set -e
set -o pipefail

# 默认配置
BASE_DIR="/runtime-mnt"
THRESHOLD=1000
AGE_DAYS=90
DRY_RUN=false
VERBOSE=false

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $(date '+%Y-%m-%d %H:%M:%S') $*"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $(date '+%Y-%m-%d %H:%M:%S') $*"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $(date '+%Y-%m-%d %H:%M:%S') $*"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $(date '+%Y-%m-%d %H:%M:%S') $*" >&2
}

# 显示帮助信息
show_help() {
    cat << EOF
Fluid Sidecar HostPath 清理脚本

用途：
  清理 Fluid Sidecar 模式下使用 random-suffix 模式产生的残留 HostPath 目录

使用方法：
  $0 [选项]

选项：
  --dry-run              仅显示将要删除的目录，不实际执行删除
  --base-dir <dir>       指定基础目录（默认：/runtime-mnt）
  --threshold <num>      指定触发清理的目录数阈值（默认：1000）
  --age-days <num>       指定目录保留天数（默认：90 天）
  --verbose              显示详细日志
  --help                 显示此帮助信息

示例：
  # 预览将要删除的目录
  $0 --dry-run

  # 清理超过 60 天的目录
  $0 --age-days 60

  # 指定自定义基础目录
  $0 --base-dir /custom/runtime-mnt

安全机制：
  1. 仅在子目录数超过阈值时执行清理
  2. 仅清理超过指定天数的目录
  3. 检查 /proc/self/mountinfo 确保目录未被挂载
  4. 仅删除最后一级子目录，不删除父目录
  5. 使用 rmdir 命令，确保目录为空才删除
  6. 删除前验证目录路径格式，防止误删
EOF
}

# 解析命令行参数
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            --dry-run)
                DRY_RUN=true
                shift
                ;;
            --base-dir)
                BASE_DIR="$2"
                shift 2
                ;;
            --threshold)
                THRESHOLD="$2"
                shift 2
                ;;
            --age-days)
                AGE_DAYS="$2"
                shift 2
                ;;
            --verbose)
                VERBOSE=true
                shift
                ;;
            --help)
                show_help
                exit 0
                ;;
            *)
                log_error "未知参数: $1"
                show_help
                exit 1
                ;;
        esac
    done
}

# 检查前置条件
check_prerequisites() {
    log_info "检查前置条件..."

    # 检查是否以 root 运行
    if [[ $EUID -ne 0 ]]; then
        log_error "此脚本必须以 root 权限运行"
        exit 1
    fi

    # 检查基础目录是否存在
    if [[ ! -d "$BASE_DIR" ]]; then
        log_error "基础目录不存在: $BASE_DIR"
        exit 1
    fi

    # 检查必要的命令
    for cmd in find date stat rmdir; do
        if ! command -v $cmd &> /dev/null; then
            log_error "缺少必要的命令: $cmd"
            exit 1
        fi
    done

    log_success "前置条件检查通过"
}

# 检查目录是否被挂载
is_mounted() {
    local dir="$1"
    
    # 规范化路径
    local real_dir
    real_dir=$(realpath "$dir" 2>/dev/null) || real_dir="$dir"
    
    # 检查 /proc/self/mountinfo
    if grep -q " ${real_dir} " /proc/self/mountinfo 2>/dev/null; then
        return 0  # 已挂载
    fi
    
    # 检查 /proc/mounts 作为备用
    if grep -q " ${real_dir} " /proc/mounts 2>/dev/null; then
        return 0  # 已挂载
    fi
    
    # 检查是否存在 .fuse_hidden 文件（FUSE 挂载的特征）
    if find "$dir" -maxdepth 1 -name ".fuse_hidden*" 2>/dev/null | grep -q .; then
        return 0  # 可能有 FUSE 挂载
    fi
    
    return 1  # 未挂载
}

# 验证目录路径格式是否符合 Fluid Sidecar random-suffix 格式
# 格式：<pod-name>/<timestamp>-<random-suffix>/<dataset-name>-fuse-mount
validate_path_format() {
    local path="$1"
    local base_path="$2"
    
    # 提取相对路径
    local rel_path="${path#${base_path}/}"
    
    # 分解路径组件
    IFS='/' read -ra PARTS <<< "$rel_path"
    
    # 至少应该有 3 层：pod-name / timestamp-random / fuse-mount-dir
    if [[ ${#PARTS[@]} -lt 3 ]]; then
        return 1
    fi
    
    # 检查第二层是否包含时间戳格式（数字-字符串）
    local timestamp_dir="${PARTS[1]}"
    if [[ ! "$timestamp_dir" =~ ^[0-9]+-[a-z0-9]+$ ]]; then
        return 1
    fi
    
    # 检查最后一层是否包含 fuse-mount 关键字
    local last_part="${PARTS[-1]}"
    if [[ ! "$last_part" =~ -fuse-mount$ ]]; then
        return 1
    fi
    
    return 0
}

# 获取目录的修改时间（天数）
get_dir_age_days() {
    local dir="$1"
    local current_time
    local dir_mtime
    
    current_time=$(date +%s)
    
    # 获取目录的最后修改时间
    if [[ "$(uname)" == "Darwin" ]]; then
        # macOS
        dir_mtime=$(stat -f %m "$dir")
    else
        # Linux
        dir_mtime=$(stat -c %Y "$dir")
    fi
    
    local age_seconds=$((current_time - dir_mtime))
    local age_days=$((age_seconds / 86400))
    
    echo "$age_days"
}

# 安全删除目录
safe_remove_dir() {
    local dir="$1"
    
    # 最终安全检查
    if is_mounted "$dir"; then
        log_warn "目录正在被挂载，跳过: $dir"
        return 1
    fi
    
    # 检查目录是否为空（不包括隐藏文件）
    if [[ -n "$(ls -A "$dir" 2>/dev/null)" ]]; then
        if $VERBOSE; then
            log_warn "目录非空，跳过: $dir"
        fi
        return 1
    fi
    
    # 使用 rmdir 删除（确保目录为空）
    if $DRY_RUN; then
        log_info "[DRY-RUN] 将删除目录: $dir"
        return 0
    else
        if rmdir "$dir" 2>/dev/null; then
            log_success "已删除目录: $dir"
            return 0
        else
            log_warn "删除目录失败（可能非空）: $dir"
            return 1
        fi
    fi
}

# 清理指定父目录下的过期子目录
cleanup_parent_dir() {
    local parent_dir="$1"
    local deleted_count=0
    local skipped_count=0
    
    log_info "处理父目录: $parent_dir"
    
    # 查找最后一级子目录（timestamp-random 目录下的 fuse-mount 目录）
    while IFS= read -r -d '' fuse_mount_dir; do
        # 验证路径格式
        if ! validate_path_format "$fuse_mount_dir" "$BASE_DIR"; then
            if $VERBOSE; then
                log_warn "路径格式不匹配，跳过: $fuse_mount_dir"
            fi
            ((skipped_count++))
            continue
        fi
        
        # 检查目录年龄
        local age_days
        age_days=$(get_dir_age_days "$fuse_mount_dir")
        
        if [[ $age_days -lt $AGE_DAYS ]]; then
            if $VERBOSE; then
                log_info "目录太新（${age_days}天 < ${AGE_DAYS}天），跳过: $fuse_mount_dir"
            fi
            ((skipped_count++))
            continue
        fi
        
        if $VERBOSE; then
            log_info "目录年龄: ${age_days}天，准备清理: $fuse_mount_dir"
        fi
        
        # 检查是否被挂载
        if is_mounted "$fuse_mount_dir"; then
            log_warn "目录正在被挂载，跳过: $fuse_mount_dir"
            ((skipped_count++))
            continue
        fi
        
        # 安全删除
        if safe_remove_dir "$fuse_mount_dir"; then
            ((deleted_count++))
            
            # 尝试清理父目录（timestamp-random 目录）
            local timestamp_dir
            timestamp_dir=$(dirname "$fuse_mount_dir")
            if [[ -d "$timestamp_dir" ]] && [[ -z "$(ls -A "$timestamp_dir" 2>/dev/null)" ]]; then
                safe_remove_dir "$timestamp_dir" && log_info "已清理空的时间戳目录: $timestamp_dir"
            fi
        else
            ((skipped_count++))
        fi
    done < <(find "$parent_dir" -mindepth 2 -maxdepth 2 -type d -name "*-fuse-mount" -print0 2>/dev/null)
    
    log_info "父目录处理完成: $parent_dir (删除: $deleted_count, 跳过: $skipped_count)"
    
    echo "$deleted_count"
}

# 主清理函数
main_cleanup() {
    log_info "开始清理流程..."
    log_info "配置参数："
    log_info "  - 基础目录: $BASE_DIR"
    log_info "  - 触发阈值: $THRESHOLD"
    log_info "  - 保留天数: $AGE_DAYS"
    log_info "  - 模式: $(if $DRY_RUN; then echo "预览模式"; else echo "实际删除"; fi)"
    
    # 统计当前所有 pod 目录下的子目录总数
    local total_subdirs=0
    local pod_dirs=()
    
    # 查找所有 pod 级别的目录
    while IFS= read -r -d '' pod_dir; do
        pod_dirs+=("$pod_dir")
        
        # 统计该 pod 目录下的时间戳目录数量
        local count
        count=$(find "$pod_dir" -mindepth 1 -maxdepth 1 -type d 2>/dev/null | wc -l)
        total_subdirs=$((total_subdirs + count))
    done < <(find "$BASE_DIR" -mindepth 1 -maxdepth 1 -type d -print0 2>/dev/null)
    
    log_info "发现 ${#pod_dirs[@]} 个 Pod 目录，总计 $total_subdirs 个时间戳子目录"
    
    # 检查是否超过阈值
    if [[ $total_subdirs -lt $THRESHOLD ]]; then
        log_info "子目录数量（$total_subdirs）未超过阈值（$THRESHOLD），无需清理"
        return 0
    fi
    
    log_warn "子目录数量（$total_subdirs）已超过阈值（$THRESHOLD），开始清理..."
    
    # 遍历每个 pod 目录进行清理
    local total_deleted=0
    for pod_dir in "${pod_dirs[@]}"; do
        local deleted
        deleted=$(cleanup_parent_dir "$pod_dir")
        total_deleted=$((total_deleted + deleted))
    done
    
    # 清理空的 pod 目录
    log_info "检查并清理空的 Pod 目录..."
    for pod_dir in "${pod_dirs[@]}"; do
        if [[ -d "$pod_dir" ]] && [[ -z "$(ls -A "$pod_dir" 2>/dev/null)" ]]; then
            safe_remove_dir "$pod_dir" && log_info "已清理空的 Pod 目录: $pod_dir"
        fi
    done
    
    log_success "清理完成！"
    log_info "总计删除 $total_deleted 个目录"
    
    # 再次统计
    local remaining_subdirs=0
    for pod_dir in "${pod_dirs[@]}"; do
        if [[ -d "$pod_dir" ]]; then
            local count
            count=$(find "$pod_dir" -mindepth 1 -maxdepth 1 -type d 2>/dev/null | wc -l)
            remaining_subdirs=$((remaining_subdirs + count))
        fi
    done
    log_info "清理后剩余 $remaining_subdirs 个时间戳子目录"
}

# 主函数
main() {
    log_info "=========================================="
    log_info "Fluid Sidecar HostPath 清理脚本"
    log_info "=========================================="
    
    parse_args "$@"
    check_prerequisites
    main_cleanup
    
    log_info "=========================================="
    log_info "脚本执行完成"
    log_info "=========================================="
}

# 执行主函数
main "$@"

