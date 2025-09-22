/**
 * 根据运行时类型生成StatefulSet名称的工具函数
 * 
 * 不同的运行时类型有不同的StatefulSet命名规则：
 * - JindoRuntime: ${runtimeName}-jindofs-${component}
 * - 其他运行时: ${runtimeName}-${component}
 */

/**
 * 根据运行时类型生成StatefulSet名称
 * @param runtimeName 运行时名称
 * @param runtimeType 运行时类型 (如 'JindoRuntime', 'AlluxioRuntime' 等)
 * @param component 组件类型 ('master' 或 'worker')
 * @returns StatefulSet名称
 */
export const generateStatefulSetName = (
  runtimeName: string, 
  runtimeType: string, 
  component: 'master' | 'worker'
): string => {
  if (runtimeType === 'JindoRuntime') {
    return `${runtimeName}-jindofs-${component}`;
  }
  return `${runtimeName}-${component}`;
};

/**
 * 根据运行时显示名称生成StatefulSet名称
 * @param runtimeName 运行时名称
 * @param runtimeDisplayName 运行时显示名称 (如 'Jindo', 'Alluxio' 等)
 * @param component 组件类型 ('master' 或 'worker')
 * @returns StatefulSet名称
 */
export const generateStatefulSetNameByDisplayName = (
  runtimeName: string, 
  runtimeDisplayName: string, 
  component: 'master' | 'worker'
): string => {
  if (runtimeDisplayName === 'Jindo') {
    return `${runtimeName}-jindofs-${component}`;
  }
  return `${runtimeName}-${component}`;
};
