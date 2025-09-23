/*
 * DataLoad ResourceStatus component
 */

import React from 'react';
import { StatusIndicator, useCacheStore as useStore } from '@ks-console/shared';
import { Card } from '@kubed/components';
import { get } from 'lodash';
import styled from 'styled-components';
import { Book2Duotone, DownloadDuotone, PlayDuotone } from '@kubed/icons';
import { getCurrentClusterFromUrl } from '../../../../utils';
import { 
  CardWrapper, 
  InfoGrid, 
  InfoItem, 
  InfoLabel, 
  InfoValue 
} from '../../../shared/components/ResourceStatusStyles';
import { getStatusIndicatorType } from '../../../../utils/getStatusIndicatorType';

// 全局t函数声明
declare const t: (key: string, options?: any) => string;

// 拓扑图样式组件（复用dataset的样式）
const TopologyContainer = styled.div`
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 24px;
  padding: 24px;
  background-color: #f9fbfd;
  border-radius: 8px;
  min-height: 120px;

  @media (max-width: 768px) {
    flex-direction: column;
    gap: 16px;
  }
`;

const TopologyNode = styled.div<{ clickable?: boolean }>`
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  padding: 16px;
  background-color: #fff;
  border: 2px solid #e9e9e9;
  border-radius: 8px;
  min-width: 100px;
  transition: all 0.2s ease;

  ${props => props.clickable && `
    cursor: pointer;

    &:hover {
      border-color: #369a6a;
      box-shadow: 0 2px 8px rgba(54, 154, 106, 0.15);
      transform: translateY(-2px);
    }
  `}
`;

const TopologyIcon = styled.div`
  font-size: 24px;
  color: #369a6a;
`;

const TopologyLabel = styled.div`
  font-size: 12px;
  font-weight: 600;
  color: #242e42;
  text-align: center;
`;

const TopologyName = styled.div`
  font-size: 14px;
  color: #242e42;
  text-align: center;
  word-break: break-all;
`;

const TopologyArrow = styled.div`
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;

  @media (max-width: 768px) {
    transform: rotate(90deg);
  }
`;

const ArrowIcon = styled.div`
  font-size: 20px;
  color: #369a6a;
`;

const ArrowLabel = styled.div`
  font-size: 10px;
  color: #79879c;
  text-align: center;
  white-space: nowrap;
`;

const ResourceStatus = () => {
  const [props] = useStore('DataLoadDetailProps');
  const { detail } = props;

  if (!detail) {
    return <div>Loading dataload details...</div>;
  }

  const currentCluster = getCurrentClusterFromUrl();

  // 处理Dataset点击跳转
  const handleDatasetClick = () => {
    const datasetName = get(detail, 'spec.dataset.name');
    const datasetNamespace = get(detail, 'spec.dataset.namespace', get(detail, 'metadata.namespace'));
    if (datasetName && datasetNamespace) {
      const url = `/fluid/${currentCluster}/${datasetNamespace}/datasets/${datasetName}/resource-status`;
      window.open(url, '_blank');
    }
  };

  // 处理任务点击跳转
  const handleJobClick = () => {
    const dataloadName = get(detail, 'metadata.name');
    const namespace = get(detail, 'metadata.namespace', 'default');
    if (dataloadName) {
      const jobName = `${dataloadName}-loader-job`;
      const url = `/clusters/${currentCluster}/projects/${namespace}/jobs/${jobName}/records`;
      window.open(url, '_blank');
    }
  };

  // 渲染拓扑图
  const renderTopologyGraph = () => {
    const dataloadName = detail.metadata?.name || 'dataload';
    const datasetName = get(detail, 'spec.dataset.name', 'unknown');
    const jobName = `${dataloadName}-loader-job`;

    return (
      <TopologyContainer>
        {/* Dataset节点 */}
        <TopologyNode clickable onClick={handleDatasetClick}>
          <TopologyIcon>
            <Book2Duotone size={24} />
          </TopologyIcon>
          <TopologyLabel>Dataset</TopologyLabel>
          <TopologyName>{datasetName}</TopologyName>
        </TopologyNode>

        {/* 箭头 */}
        <TopologyArrow>
          <ArrowIcon>→</ArrowIcon>
          <ArrowLabel>{t('LOADS_FROM')}</ArrowLabel>
        </TopologyArrow>

        {/* DataLoad节点 */}
        <TopologyNode>
          <TopologyIcon>
            <DownloadDuotone size={24} />
          </TopologyIcon>
          <TopologyLabel>DataLoad</TopologyLabel>
          <TopologyName>{dataloadName}</TopologyName>
        </TopologyNode>

        {/* 箭头 */}
        <TopologyArrow>
          <ArrowIcon>→</ArrowIcon>
          <ArrowLabel>{t('CREATES')}</ArrowLabel>
        </TopologyArrow>

        {/* 任务节点 */}
        <TopologyNode clickable onClick={handleJobClick}>
          <TopologyIcon>
            <PlayDuotone size={24} />
          </TopologyIcon>
          <TopologyLabel>Job</TopologyLabel>
          <TopologyName>{jobName}</TopologyName>
        </TopologyNode>
      </TopologyContainer>
    );
  };

  return (
    <>
      {/* 基本信息卡片 */}
      <CardWrapper>
        <Card sectionTitle={t('BASIC_INFORMATION')}>
          <InfoGrid>
            <InfoItem>
              <InfoLabel>{t('STATUS')}</InfoLabel>
              <InfoValue>
                <StatusIndicator
                    type={getStatusIndicatorType(get(detail, 'status.phase', '-'))}
                    motion={false}
                >
                  {get(detail, 'status.phase', '-')}
                </StatusIndicator>
                
                </InfoValue>
            </InfoItem>
            <InfoItem>
              <InfoLabel>{t('DATASET')}</InfoLabel>
              <InfoValue>{get(detail, 'spec.dataset.name', '-')}</InfoValue>
            </InfoItem>
            <InfoItem>
              <InfoLabel>{t('POLICY')}</InfoLabel>
              <InfoValue>{get(detail, 'spec.policy', 'Once')}</InfoValue>
            </InfoItem>
            <InfoItem>
              <InfoLabel>{t('LOAD_METADATA')}</InfoLabel>
              <InfoValue>{get(detail, 'spec.loadMetadata', false) ? t('TRUE') : t('FALSE')}</InfoValue>
            </InfoItem>
            <InfoItem>
              <InfoLabel>{t('DURATION')}</InfoLabel>
              <InfoValue>{get(detail, 'status.duration', '-')}</InfoValue>
            </InfoItem>
            <InfoItem>
              <InfoLabel>{t('CREATION_TIME')}</InfoLabel>
              <InfoValue>{get(detail, 'metadata.creationTimestamp', '-')}</InfoValue>
            </InfoItem>
          </InfoGrid>
        </Card>
      </CardWrapper>

      {/* 拓扑图卡片 */}
      <CardWrapper>
        <Card sectionTitle={t('DATALOAD_TOPOLOGY')}>
          {renderTopologyGraph()}
        </Card>
      </CardWrapper>

      {/* 目标路径信息 */}
      {get(detail, 'spec.target') && (
        <CardWrapper>
          <Card sectionTitle={t('TARGET_PATHS')}>
            <InfoGrid>
              {get(detail, 'spec.target', []).map((target: any, index: number) => (
                <React.Fragment key={index}>
                  <InfoItem>
                    <InfoLabel>{t('PATH')} {index + 1}</InfoLabel>
                    <InfoValue>{target.path || '-'}</InfoValue>
                  </InfoItem>
                  <InfoItem>
                    <InfoLabel>{t('REPLICAS')}</InfoLabel>
                    <InfoValue>{target.replicas || '-'}</InfoValue>
                  </InfoItem>
                </React.Fragment>
              ))}
            </InfoGrid>
          </Card>
        </CardWrapper>
      )}

    </>
  );
};

export default ResourceStatus;
