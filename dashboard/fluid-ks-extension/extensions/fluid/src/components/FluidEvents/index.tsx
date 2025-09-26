/*
 * Fluid Events component - 公共事件组件
 * 用于数据集和运行时详情页的事件显示
 */

import React from 'react';
import { Events } from '@ks-console/shared';
import { get } from 'lodash';
import { getCurrentClusterFromUrl } from '../../utils';

// 数据集状态判断
const getDatasetPhase = (item: any) => {
  const deletionTime = get(item, 'metadata.deletionTimestamp');
  if (deletionTime) {
    return 'Terminating';
  }
  return get(item, 'status.phase', 'Unknown');
};

// 运行时状态判断
const getRuntimePhase = (item: any) => {
  const deletionTime = get(item, 'metadata.deletionTimestamp');
  if (deletionTime) {
    return 'Terminating';
  }
  // 运行时的状态优先级：master > worker > fuse
  return get(item, 'status.masterPhase', 
    get(item, 'status.workerPhase', 
      get(item, 'status.fusePhase', 'Unknown')));
};

// DataLoad状态判断
const getDataLoadPhase = (item: any) => {
  const deletionTime = get(item, 'metadata.deletionTimestamp');
  if (deletionTime) {
    return 'Terminating';
  }
  return get(item, 'status.phase', 'Unknown');
};

// 获取资源状态的通用函数
const getResourcePhase = (item: any, resourceType: 'dataset' | 'runtime' | 'dataload') => {
  switch (resourceType) {
    case 'dataset':
      return getDatasetPhase(item);
    case 'runtime':
      return getRuntimePhase(item);
    case 'dataload':
      return getDataLoadPhase(item);
    default:
      return get(item, 'status.phase', 'Unknown');
  }
};

interface FluidEventsProps {
  /** 原始详情数据 */
  detail: any;
  /** 模块名称，用于传递给Events组件 */
  module: string;
  /** 资源类型，用于状态判断 */
  resourceType: 'dataset' | 'runtime' | 'dataload';
  /** 加载状态文本 */
  loadingText?: string;
}

const FluidEvents: React.FC<FluidEventsProps> = ({
  detail: rawDetail,
  module,
  resourceType,
  loadingText = 'Loading details...'
}) => {
  console.log(`FluidEvents[${resourceType}] - 原始detail:`, rawDetail);
  console.log(`FluidEvents[${resourceType}] - module:`, module);

  // 如果没有原始数据，显示加载状态
  if (!rawDetail) {
    return <div>{loadingText}</div>;
  }

  // 获取集群信息
  const cluster = getCurrentClusterFromUrl();

  // 规范化数据，转换为Events组件期望的格式
  const normalizedDetail = {
    // 核心标识字段（Events组件必需）
    uid: get(rawDetail, 'metadata.uid'),
    name: get(rawDetail, 'metadata.name'),
    namespace: get(rawDetail, 'metadata.namespace'),
    cluster: cluster,

    // 状态相关字段
    phase: getResourcePhase(rawDetail, resourceType),
    deletionTime: get(rawDetail, 'metadata.deletionTimestamp'),

    // 原始数据引用（组件内部可能依赖）
    _originData: rawDetail,

    // 其他辅助字段
    creationTime: get(rawDetail, 'metadata.creationTimestamp'),
    kind: get(rawDetail, 'kind',
      resourceType === 'dataset' ? 'Dataset' :
      resourceType === 'runtime' ? 'Runtime' : 'DataLoad'),
    apiVersion: get(rawDetail, 'apiVersion', 'data.fluid.io/v1alpha1'),

    // 保持原有的metadata结构以防某些组件依赖
    metadata: rawDetail.metadata,
  };

  console.log(`FluidEvents[${resourceType}] - 规范化后的detail:`, normalizedDetail);

  return (
    <Events detail={normalizedDetail} module={module} />
  );
};

export default FluidEvents;
