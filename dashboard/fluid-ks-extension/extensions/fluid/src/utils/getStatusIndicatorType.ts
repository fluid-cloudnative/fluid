// 获取状态指示器类型函数
export const getStatusIndicatorType = (phase: string): 'success' | 'warning' | 'error' | 'default' => {
  switch (phase?.toLowerCase()) {
    case 'ready':
    case 'running':
    case 'bound':
    case 'executing':
      return 'success';
    case 'creating':
    case 'updating':
    case 'pending':
    case 'complete':
    case 'notbound':
      return 'warning';
    case 'failed':
    case 'error':
    case 'terminating':
      return 'error';
    default:
      return 'default';
  }
};