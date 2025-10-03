import { Mount } from './types';

/**
 * 将 options 数组转换为对象格式
 */
export const convertOptionsToObject = (options: Array<{ key: string; value: string }>): Record<string, string> => {
  return options.reduce((acc, { key, value }) => {
    // 只有当key和value都不为空时才添加到最终的options对象中
    if (key.trim() && value.trim()) {
      acc[key.trim()] = value.trim();
    }
    return acc;
  }, {} as Record<string, string>);
};

/**
 * 将 options 对象转换为数组格式
 */
export const convertOptionsToArray = (options?: Record<string, string>): Array<{ key: string; value: string }> => {
  if (!options) return [];
  return Object.entries(options).map(([key, value]) => ({ key, value }));
};

/**
 * 转换挂载点数据为表单提交格式
 */
export const convertMountsForSubmission = (mounts: Mount[]) => {
  return mounts.map(mount => ({
    mountPoint: mount.mountPoint,
    name: mount.name,
    path: mount.path,
    readOnly: mount.readOnly,
    shared: mount.shared,
    options: convertOptionsToObject(mount.options),
    encryptOptions: mount.encryptOptions?.filter(option =>
      option.name.trim() &&
      option.valueFrom?.secretKeyRef.name.trim() &&
      option.valueFrom?.secretKeyRef.key.trim()
    ).map(option => ({
      name: option.name.trim(),
      valueFrom: {
        secretKeyRef: {
          name: option.valueFrom!.secretKeyRef.name.trim(),
          key: option.valueFrom!.secretKeyRef.key.trim()
        }
      }
    })),
  }));
};

/**
 * 从表单数据初始化挂载点
 */
export const initializeMountsFromFormData = (formMounts?: any[]): Mount[] => {
  if (!formMounts || formMounts.length === 0) {
    return [{
      mountPoint: '',
      name: 'default',
      path: '',
      readOnly: false,
      shared: true,
      options: [],
      encryptOptions: [],
    }];
  }

  return formMounts.map(mount => ({
    mountPoint: mount.mountPoint,
    name: mount.name,
    path: mount.path || '',
    readOnly: mount.readOnly || false,
    shared: mount.shared !== undefined ? mount.shared : true,
    options: convertOptionsToArray(mount.options),
    encryptOptions: mount.encryptOptions || [],
  }));
};

/**
 * 生成新挂载点的默认名称
 */
export const generateMountName = (existingMounts: Mount[]): string => {
  return `mount-${existingMounts.length}`;
};
