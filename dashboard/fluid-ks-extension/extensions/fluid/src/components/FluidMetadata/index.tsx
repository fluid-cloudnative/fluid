/*
 * Fluid Metadata component - 公共元数据组件
 * 用于数据集和运行时详情页的元数据显示
 */

import React from 'react';
import { useCacheStore as useStore } from '@ks-console/shared';
import { MetaData } from '@ks-console/shared';
import { Loading } from '@kubed/components';
import { isEmpty, get } from 'lodash';
import styled from 'styled-components';

// 全局t函数声明
declare const t: (key: string, options?: any) => string;

const EmptyTip = styled.div`
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 20px 0;
  color: #79879c;
`;

interface FluidMetadataProps {
  storeKey: 'DatasetDetailProps' | 'RuntimeDetailProps' | 'DataLoadDetailProps';
  loadingText?: string;
}

const FluidMetadata: React.FC<FluidMetadataProps> = ({
  storeKey,
  loadingText = 'Loading details...'
}) => {
  const [props] = useStore(storeKey);
  const { detail, isLoading } = props;
  
  if (isLoading) {
    return <Loading className="page-loading" />;
  }
  
  // 如果没有详情数据，显示加载提示
  if (!detail) {
    return <div>{loadingText}</div>;
  }
  
  // 获取标签和注解
  const labels = get(detail, 'metadata.labels', {});
  const annotations = get(detail, 'metadata.annotations', {});
  
  // 检查是否有标签和注解数据
  const hasLabels = !isEmpty(labels);
  const hasAnnotations = !isEmpty(annotations);
  const hasMetadata = hasLabels || hasAnnotations;
  
  // 如果没有元数据，显示"无数据"提示
  if (!hasMetadata) {
    return <EmptyTip>{t('NO_RESOURCE_FOUND')}</EmptyTip>;
  }
  
  // 准备符合 MetaData 组件期望的数据结构
  const metadataDetail = {
    labels,
    annotations
  };
  
  // 使用 KubeSphere 的 MetaData 组件显示元数据
  return <MetaData detail={metadataDetail} />;
};

export default FluidMetadata;
