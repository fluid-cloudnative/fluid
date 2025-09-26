/**
 * 跨集群API请求工具类
 * 统一处理集群前缀和API路径
 * 基于URL路径获取集群信息，无需全局状态管理
 */

/**
 * 从当前URL路径中解析集群名称
 * @returns 集群名称，如 'host' 或 'member-cluster'
 */
export const getCurrentClusterFromUrl = (): string => {
  const pathSegments = window.location.pathname.split('/');
  // URL 格式: /fluid/{cluster}/...
  const fluidIndex = pathSegments.indexOf('fluid');
  if (fluidIndex >= 0 && fluidIndex + 1 < pathSegments.length) {
    return pathSegments[fluidIndex + 1];
  }
  return 'host'; // 默认集群
};

/**
 * 获取当前选择的集群前缀
 * @returns 集群前缀，如 /clusters/host 或 /clusters/member-cluster
 */
export const getClusterPrefix = (): string => {
  const cluster = getCurrentClusterFromUrl();
  return `/clusters/${cluster}`;
};

/**
 * 获取完整的API路径，自动添加集群前缀
 * @param path 原始API路径
 * @returns 包含集群前缀的完整路径
 */
export const getApiPath = (path: string): string => {
  // 如果路径已经包含集群前缀，则直接返回
  if (path.startsWith('/clusters/')) {
    return path;
  }

  // 对于WebSocket路径，也需要添加集群前缀
  if (path.startsWith('/apis/') || path.startsWith('/kapis/') || path.startsWith('/api/')) {
    return `${getClusterPrefix()}${path}`;
  }

  return `${getClusterPrefix()}${path}`;
};

/**
 * 封装的请求方法，自动处理集群前缀
 * @param path API路径
 * @param options fetch选项
 * @returns Promise<Response>
 */
export const request = async (path: string, options?: RequestInit): Promise<Response> => {
  const apiPath = getApiPath(path);
  return fetch(apiPath, options);
};

/**
 * 封装的JSON请求方法
 * @param path API路径
 * @param options fetch选项
 * @returns Promise<any>
 */
export const requestJson = async (path: string, options?: RequestInit): Promise<any> => {
  const response = await request(path, options);
  
  if (!response.ok) {
    throw new Error(`API request failed: ${response.status} ${response.statusText}`);
  }
  
  return response.json();
};

/**
 * 获取WebSocket URL，自动添加集群前缀
 * @param path WebSocket路径
 * @returns 完整的WebSocket URL
 */
export const getWebSocketUrl = (path: string): string => {
  const apiPath = getApiPath(path);
  // 将http/https协议转换为ws/wss
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  const host = window.location.host;
  return `${protocol}//${host}${apiPath}`;
};

/**
 * 封装的PATCH请求方法
 * @param path API路径
 * @param data 请求数据
 * @param options 额外的fetch选项
 * @returns Promise<any>
 */
export const requestPatch = async (path: string, data: any, options?: RequestInit): Promise<any> => {
  const response = await request(path, {
    method: 'PATCH',
    headers: {
      'Content-Type': 'application/merge-patch+json',
      ...options?.headers,
    },
    body: JSON.stringify(data),
    ...options,
  });

  if (!response.ok) {
    throw new Error(`PATCH request failed: ${response.status} ${response.statusText}`);
  }

  return response.json();
};

/**
 * 获取当前集群名称
 * @returns 当前集群名称
 * @deprecated 使用 getCurrentClusterFromUrl() 替代
 */
export const getCurrentCluster = (): string => {
  return getCurrentClusterFromUrl();
};
