/*
 * 详情页标签页配置工具函数
 * 用于生成dataset和runtime详情页的标签页配置
 */

// 全局t函数声明
declare const t: (key: string, options?: any) => string;

export interface TabConfig {
  title: string;
  path: string;
}

/**
 * 创建详情页标签页配置
 * @param cluster 集群名称
 * @param namespace 命名空间
 * @param name 资源名称
 * @param module 模块名称 ('datasets' | 'runtimes' | 'dataloads')
 * @returns 标签页配置数组
 */
export function createDetailTabs(
  cluster: string,
  namespace: string,
  name: string,
  module: 'datasets' | 'runtimes' | 'dataloads'
): TabConfig[] {
  const clusterName = cluster || 'host';
  const path = `/fluid/${clusterName}/${namespace}/${module}/${name}`;
  
  return [
    {
      title: t('RESOURCE_STATUS'),
      path: `${path}/resource-status`,
    },
    {
      title: t('METADATA'),
      path: `${path}/metadata`,
    },
    {
      title: t('EVENTS'),
      path: `${path}/events`,
    },
  ];
}
