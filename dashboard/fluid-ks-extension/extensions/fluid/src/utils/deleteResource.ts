/**
 * 通用资源删除工具
 * 支持删除 Dataset、DataLoad、Runtime 等不同类型的资源
 */

import { request } from './request';
import { notify } from '@kubed/components';

// 全局t函数声明
declare const t: (key: string, options?: any) => string;

/**
 * 删除资源的选项
 */
export interface DeleteResourceOptions {
  /** 资源类型 */
  resourceType: 'dataset' | 'dataload' | 'runtime';
  /** 资源名称 */
  name: string;
  /** 命名空间 */
  namespace: string;
  /** Runtime类型，仅当resourceType为runtime时需要 */
  runtimeType?: string;
}

/**
 * 处理删除操作的选项
 */
export interface HandleDeleteOptions extends DeleteResourceOptions {
  /** 删除成功回调 */
  onSuccess?: () => void;
  /** 删除失败回调 */
  onError?: (error: Error) => void;
  /** 自定义确认消息 */
  confirmMessage?: string;
  /** 自定义成功消息 */
  successMessage?: string;
  /** 是否跳过确认对话框 */
  skipConfirm?: boolean;
}

/**
 * 获取资源的显示名称
 */
const getResourceDisplayName = (resourceType: string): string => {
  switch (resourceType) {
    case 'dataset':
      return t('DATASET') || '数据集';
    case 'dataload':
      return t('DATALOAD') || '数据加载任务';
    case 'runtime':
      return t('RUNTIME') || '运行时';
    default:
      return resourceType;
  }
};

/**
 * 构建删除API路径
 */
const getDeleteApiPath = (options: DeleteResourceOptions): string => {
  const { resourceType, namespace, name, runtimeType } = options;
  
  switch (resourceType) {
    case 'dataset':
      return `/kapis/data.fluid.io/v1alpha1/namespaces/${namespace}/datasets/${name}`;
    case 'dataload':
      return `/kapis/data.fluid.io/v1alpha1/namespaces/${namespace}/dataloads/${name}`;
    case 'runtime':
      if (!runtimeType) {
        throw new Error('runtimeType is required for runtime resources');
      }
      // 将 AlluxioRuntime -> alluxioruntimes
      const pluralType = runtimeType.toLowerCase() + 's';
      return `/kapis/data.fluid.io/v1alpha1/namespaces/${namespace}/${pluralType}/${name}`;
    default:
      throw new Error(`Unsupported resource type: ${resourceType}`);
  }
};

/**
 * 获取默认确认消息
 */
const getDefaultConfirmMessage = (resourceType: string, name: string): string => {
  const displayName = getResourceDisplayName(resourceType);
  return `确定要删除${displayName} "${name}" 吗？此操作不可撤销。删除操作不会立马成功，请等待一会重新刷新`;
};

/**
 * 获取默认成功消息
 */
const getDefaultSuccessMessage = (resourceType: string, name: string): string => {
  const displayName = getResourceDisplayName(resourceType);
  return `成功删除${displayName} "${name}"`;
};

/**
 * 删除单个资源
 * @param options 删除选项
 * @returns Promise<boolean> 删除是否成功
 */
export const deleteResource = async (options: DeleteResourceOptions): Promise<boolean> => {
  try {
    const apiPath = getDeleteApiPath(options);
    const response = await request(apiPath, {
      method: 'DELETE',
    });

    if (!response.ok) {
      throw new Error(`删除失败: ${response.status} ${response.statusText}`);
    }

    return true;
  } catch (error) {
    console.error(`删除${getResourceDisplayName(options.resourceType)}失败:`, error);
    throw error;
  }
};

/**
 * 处理资源删除操作（包含确认对话框和通知）
 * @param options 处理选项
 */
export const handleResourceDelete = async (options: HandleDeleteOptions): Promise<void> => {
  const {
    resourceType,
    name,
    namespace,
    runtimeType,
    onSuccess,
    onError,
    confirmMessage,
    successMessage,
    skipConfirm = false
  } = options;

  // 确认对话框
  if (!skipConfirm) {
    const message = confirmMessage || getDefaultConfirmMessage(resourceType, name);
    const confirmed = window.confirm(message);
    
    if (!confirmed) {
      return;
    }
  }

  try {
    // 执行删除
    await deleteResource({
      resourceType,
      name,
      namespace,
      runtimeType
    });

    // 显示成功消息
    const message = successMessage || getDefaultSuccessMessage(resourceType, name);
    notify.success(message);

    // 执行成功回调
    onSuccess?.();
  } catch (error) {
    const errorMessage = error instanceof Error ? error.message : String(error);
    console.error('删除操作失败:', error);
    
    // 显示错误消息
    notify.error(`删除${getResourceDisplayName(resourceType)}失败: ${errorMessage}`);
    
    // 执行错误回调
    onError?.(error instanceof Error ? error : new Error(errorMessage));
  }
};

/**
 * 批量删除资源
 * @param resources 要删除的资源列表
 * @param options 删除选项
 */
export const handleBatchResourceDelete = async (
  resources: Array<{ name: string; namespace: string }>,
  options: Omit<HandleDeleteOptions, 'name' | 'namespace'>
): Promise<void> => {
  const { resourceType, runtimeType, onSuccess, onError, skipConfirm = false } = options;

  if (resources.length === 0) {
    return;
  }

  // 确认对话框
  if (!skipConfirm) {
    const displayName = getResourceDisplayName(resourceType);
    const message = `确定要删除选中的 ${resources.length} 个${displayName}吗？此操作不可撤销。删除操作不会立马成功，请等待一会重新刷新`;
    const confirmed = window.confirm(message);
    
    if (!confirmed) {
      return;
    }
  }

  try {
    // 并行删除所有资源
    const deletePromises = resources.map(resource =>
      deleteResource({
        resourceType,
        name: resource.name,
        namespace: resource.namespace,
        runtimeType
      })
    );

    await Promise.all(deletePromises);

    // 显示成功消息
    const displayName = getResourceDisplayName(resourceType);
    notify.success(`成功删除 ${resources.length} 个${displayName}`);

    // 执行成功回调
    onSuccess?.();
  } catch (error) {
    const errorMessage = error instanceof Error ? error.message : String(error);
    console.error('批量删除失败:', error);
    
    // 显示错误消息
    const displayName = getResourceDisplayName(resourceType);
    notify.error(`删除${displayName}失败: ${errorMessage}`);
    
    // 执行错误回调
    onError?.(error instanceof Error ? error : new Error(errorMessage));
  }
};
