/*
 * Dataset Resource Status component
 */

import React, { useState, useEffect } from 'react';
import { StatusIndicator, useCacheStore as useStore } from '@ks-console/shared';
import { Card } from '@kubed/components';
import { Book2Duotone, RocketDuotone, StorageDuotone, AppstoreDuotone, FolderDuotone, DatabaseSealDuotone } from '@kubed/icons';
import { get } from 'lodash';
import styled from 'styled-components';
import { SimpleCircle } from '@ks-console/shared';
import { request, getCurrentCluster } from '../../../../utils/request';
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

// 定义Mount类型
interface Mount {
  mountPoint: string;
  name?: string;
  options?: Record<string, string>;
  path?: string;
  readOnly?: boolean;
  shared?: boolean;
}

// 定义Runtime类型
interface Runtime {
  name: string;
  namespace: string;
  type: string;
  category?: string;
  masterReplicas?: number;
}



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

const MountCard = styled.div`
  border: 1px solid #e9e9e9;
  border-radius: 4px;
  padding: 12px;
  margin-bottom: 12px;
  background-color: #f9fbfd;
`;

const MountHeader = styled.div`
  display: flex;
  align-items: center;
  margin-bottom: 8px;
  padding-bottom: 8px;
  border-bottom: 1px solid #e9e9e9;
`;

const MountTitle = styled.div`
  font-weight: 600;
  font-size: 14px;
  color: #242e42;
`;

const MountDetails = styled.div`
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 8px;
  
  @media (max-width: 576px) {
    grid-template-columns: 1fr;
  }
`;

const MountItem = styled.div`
  display: flex;
`;

const MountLabel = styled.div`
  font-weight: 600;
  color: #79879c;
  margin-right: 8px;
  min-width: 80px;
  font-size: 12px;
`;

const MountValue = styled.div`
  color: #242e42;
  font-size: 12px;
`;

const ResourceStatus = () => {
  const [props] = useStore('DatasetDetailProps');
  const { detail } = props;

  // 卷检测状态
  const [volumeExists, setVolumeExists] = useState<boolean>(false);
  const [volumeLoading, setVolumeLoading] = useState<boolean>(false);

  // PVC检测状态
  const [pvcExists, setPvcExists] = useState<boolean>(false);
  const [pvcLoading, setPvcLoading] = useState<boolean>(false);

  // 解析缓存百分比
  const getCachePercentage = () => {
    const percentage = get(detail, 'status.cacheStates.cachedPercentage', '0%');
    return parseInt(percentage.replace('%', ''), 10) || 0;
  };

  // 将不同数据单位统一转换为 GiB
  const convertUnit = (value: string): number => {
    if (!value || value === '-') return 0;
    
    // 提取数字部分和单位部分
    const regex = /^([\d.]+)\s*([A-Za-z]+)$/;
    const match = value.match(regex);
    
    if (!match) return parseFloat(value) || 0;
    
    const num = parseFloat(match[1]);
    const unit = match[2].toLowerCase();
    
    // 转换为 GiB
    switch (unit) {
      case 'b':
        return num / (1024 * 1024 * 1024);
      case 'kb':
      case 'kib':
        return num / (1024 * 1024);
      case 'mb':
      case 'mib':
        return num / 1024;
      case 'gb':
      case 'gib':
        return num;
      case 'tb':
      case 'tib':
        return num * 1024;
      case 'pb':
      case 'pib':
        return num * 1024 * 1024;
      default:
        return num;
    }
  };

  // 检测卷是否存在
  const checkVolumeExists = async (namespace: string, datasetName: string) => {
    try {
      setVolumeLoading(true);
      const volumeName = `${namespace}-${datasetName}`;
      const response = await request('/api/v1/persistentvolumes');

      if (!response.ok) {
        throw new Error(`Failed to fetch persistent volumes: ${response.statusText}`);
      }

      const data = await response.json();
      const exists = data.items?.some((pv: any) => pv.metadata?.name === volumeName) || false;
      setVolumeExists(exists);
    } catch (error) {
      console.error('Failed to check volume existence:', error);
      setVolumeExists(false);
    } finally {
      setVolumeLoading(false);
    }
  };

  // 检测PVC是否存在
  const checkPvcExists = async (namespace: string, datasetName: string) => {
    try {
      setPvcLoading(true);
      const response = await request(`/api/v1/namespaces/${namespace}/persistentvolumeclaims`);

      if (!response.ok) {
        throw new Error(`Failed to fetch persistent volume claims: ${response.statusText}`);
      }

      const data = await response.json();
      const exists = data.items?.some((pvc: any) => pvc.metadata?.name === datasetName) || false;
      setPvcExists(exists);
    } catch (error) {
      console.error('Failed to check PVC existence:', error);
      setPvcExists(false);
    } finally {
      setPvcLoading(false);
    }
  };

  // 监听数据集变化，检测卷和PVC
  useEffect(() => {
    const namespace = get(detail, 'metadata.namespace');
    const datasetName = get(detail, 'metadata.name');

    if (namespace && datasetName) {
      checkVolumeExists(namespace, datasetName);
      checkPvcExists(namespace, datasetName);

      // 设置定时检测
      const interval = setInterval(() => {
        checkVolumeExists(namespace, datasetName);
        checkPvcExists(namespace, datasetName);
      }, 30000); // 每30秒检测一次

      return () => clearInterval(interval);
    }
  }, [detail]);

  // 处理运行时点击
  const handleRuntimeClick = (runtime: Runtime) => {
    // 导航到运行时详情页 - 在新窗口中打开，包含集群信息
    const currentCluster = getCurrentCluster();
    const namespace = get(detail, 'metadata.namespace');
    const url = `/fluid/${currentCluster}/${namespace}/runtimes/${runtime.name}/resource-status`;
    console.log('Opening runtime in new window:', runtime.name, 'in namespace:', namespace, 'cluster:', currentCluster);
    window.open(url, '_blank');
  };

  // 处理卷点击
  const handleVolumeClick = () => {
    const currentCluster = getCurrentCluster();
    const namespace = get(detail, 'metadata.namespace');
    const datasetName = get(detail, 'metadata.name');
    const volumeName = `${namespace}-${datasetName}`;
    const url = `/clusters/${currentCluster}/pv/${volumeName}/resource-status`;
    console.log('Opening volume in new window:', volumeName, 'cluster:', currentCluster);
    window.open(url, '_blank');
  };

  // 处理PVC点击
  const handlePVCClick = () => {
    const currentCluster = getCurrentCluster();
    const namespace = get(detail, 'metadata.namespace');
    const datasetName = get(detail, 'metadata.name');
    // 根据提供的路径格式: /clusters/host/projects/{namespace}/volumes/{pvcName}/resource-status
    const url = `/clusters/${currentCluster}/projects/${namespace}/volumes/${datasetName}/resource-status`;
    console.log('Opening PVC in new window:', datasetName, 'in namespace:', namespace,'cluster:', currentCluster);
    window.open(url, '_blank');
  };

  // 新的拓扑图可视化
  const renderTopologyGraph = () => {
    const datasetName = detail.metadata?.name || 'dataset';
    const namespace = detail.metadata?.namespace || 'default';
    const runtimes = get(detail, 'status.runtimes', []) as Runtime[];
    const mounts = get(detail, 'spec.mounts', []) as Mount[];
    const volumeName = `${namespace}-${datasetName}`;
    console.log("detail",detail)

    return (
      <TopologyContainer>
        {/* 数据源节点 */}
        {mounts.length > 0 && (
          <>
            <TopologyNode>
              <TopologyIcon>
                <FolderDuotone size={24} />
              </TopologyIcon>
              <TopologyLabel>{t('DATA_SOURCE')}</TopologyLabel>
              <TopologyName>{mounts[0].mountPoint}</TopologyName>
            </TopologyNode>
            <TopologyArrow>
              <ArrowIcon>→</ArrowIcon>
              <ArrowLabel>{t('CONNECTED_TO')}</ArrowLabel>
            </TopologyArrow>
          </>
        )}

        {/* 数据集节点 */}
        <TopologyNode>
          <TopologyIcon>
            <Book2Duotone size={24} />
          </TopologyIcon>
          <TopologyLabel>Dataset</TopologyLabel>
          <TopologyName>{datasetName}</TopologyName>
        </TopologyNode>

        {/* 运行时节点 */}
        {runtimes.length > 0 ? (
          <>
            <TopologyArrow>
              <ArrowIcon>→</ArrowIcon>
              <ArrowLabel>{t('MANAGED_BY')}</ArrowLabel>
            </TopologyArrow>
            <TopologyNode clickable onClick={() => handleRuntimeClick(runtimes[0])}>
              <TopologyIcon>
                <RocketDuotone size={24} />
              </TopologyIcon>
              <TopologyLabel>Runtime ({runtimes[0].type})</TopologyLabel>
              <TopologyName>{runtimes[0].name}</TopologyName>
            </TopologyNode>
          </>
        ) : (
          <>
            <TopologyArrow>
              <ArrowIcon>→</ArrowIcon>
              <ArrowLabel>{t('MANAGED_BY')}</ArrowLabel>
            </TopologyArrow>
            <TopologyNode>
              <TopologyIcon>
                <RocketDuotone size={24} />
              </TopologyIcon>
              <TopologyLabel>Runtime</TopologyLabel>
              <TopologyName>未配置</TopologyName>
            </TopologyNode>
          </>
        )}

        {/* PVC节点 */}
        {pvcLoading ? (
          <>
            <TopologyArrow>
              <ArrowIcon>→</ArrowIcon>
              <ArrowLabel>{t('CREATES')}</ArrowLabel>
            </TopologyArrow>
            <TopologyNode>
              <TopologyIcon>
                <DatabaseSealDuotone size={24} />
              </TopologyIcon>
              <TopologyLabel>PVC</TopologyLabel>
              <TopologyName>检测中...</TopologyName>
            </TopologyNode>
          </>
        ) : pvcExists ? (
          <>
            <TopologyArrow>
              <ArrowIcon>→</ArrowIcon>
              <ArrowLabel>{t('CREATES')}</ArrowLabel>
            </TopologyArrow>
            <TopologyNode clickable onClick={handlePVCClick}>
              <TopologyIcon>
                <DatabaseSealDuotone size={24} />
              </TopologyIcon>
              <TopologyLabel>PVC</TopologyLabel>
              <TopologyName>{datasetName}</TopologyName>
            </TopologyNode>
          </>
        ) : (
          <>
            <TopologyArrow>
              <ArrowIcon>→</ArrowIcon>
              <ArrowLabel>{t('CREATES')}</ArrowLabel>
            </TopologyArrow>
            <TopologyNode>
              <TopologyIcon>
                <DatabaseSealDuotone size={24} />
              </TopologyIcon>
              <TopologyLabel>PVC</TopologyLabel>
              <TopologyName>未创建</TopologyName>
            </TopologyNode>
          </>
        )}

        {/* 卷节点 */}
        {volumeLoading ? (
          <>
            <TopologyArrow>
              <ArrowIcon>→</ArrowIcon>
              <ArrowLabel>{t('MOUNTED_AS')}</ArrowLabel>
            </TopologyArrow>
            <TopologyNode>
              <TopologyIcon>
                <StorageDuotone size={24} />
              </TopologyIcon>
              <TopologyLabel>Volume</TopologyLabel>
              <TopologyName>检测中...</TopologyName>
            </TopologyNode>
          </>
        ) : volumeExists ? (
          <>
            <TopologyArrow>
              <ArrowIcon>→</ArrowIcon>
              <ArrowLabel>{t('MOUNTED_AS')}</ArrowLabel>
            </TopologyArrow>
            <TopologyNode clickable onClick={handleVolumeClick}>
              <TopologyIcon>
                <StorageDuotone size={24} />
              </TopologyIcon>
              <TopologyLabel>Volume</TopologyLabel>
              <TopologyName>{volumeName}</TopologyName>
            </TopologyNode>
          </>
        ) : null}

        {/* 应用节点 */}
        <TopologyArrow>
          <ArrowIcon>→</ArrowIcon>
          <ArrowLabel>{t('CONSUME')}</ArrowLabel>
        </TopologyArrow>
        <TopologyNode>
          <TopologyIcon>
            <AppstoreDuotone size={24} />
          </TopologyIcon>
          <TopologyLabel>{t('APPLICATION')}</TopologyLabel>
          <TopologyName>{t('VARIOUS_PODS')}</TopologyName>
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
              <StatusIndicator
                    type={getStatusIndicatorType(get(detail, 'status.phase', '-'))}
                    motion={false}
              >
                  <InfoValue>{get(detail, 'status.phase', '-')}</InfoValue>
              </StatusIndicator>
              
            </InfoItem>
            <InfoItem>
              <InfoLabel>{t('TOTAL_FILES')}</InfoLabel>
              <InfoValue>{get(detail, 'status.fileNum', '-')}</InfoValue>
            </InfoItem>
            <InfoItem>
              <InfoLabel>{t('UFS_TOTAL')}</InfoLabel>
              <InfoValue>{get(detail, 'status.ufsTotal', '-')}</InfoValue>
            </InfoItem>
          </InfoGrid>
        </Card>
      </CardWrapper>
      
      {/* 缓存状态卡片 */}
      <CardWrapper>
        <Card sectionTitle={t('CACHE_STATUS')}>
          <InfoGrid>
            <InfoItem>
              <InfoLabel>{t('CACHE_CAPACITY_USAGE')}</InfoLabel>
              <SimpleCircle
                theme="light"
                title={t('CACHE_CAPACITY_USAGE')}
                categories={[t('CACHED'), t('CACHE_CAPACITY')]}
                value={convertUnit(get(detail, 'status.cacheStates.cached', '-'))}
                total={convertUnit(get(detail, 'status.cacheStates.cacheCapacity', '-'))}
                showRate
                unit='GiB'
              />
            </InfoItem>
            <InfoItem>
              <InfoLabel>{t('CACHE_HIT_RATIO')}</InfoLabel>
              <SimpleCircle
                theme="light"
                title={t('CACHE_HIT_RATIO')}
                categories={[t('CACHE_HIT_RATIO')]}
                value={(() => {
                  const ratio = get(detail, 'status.cacheStates.cacheHitRatio', 0);
                  if (ratio === '-' || ratio === null || ratio === undefined) return 0;
                  // 如果是小数形式(0-1)，转换为百分比(0-100)
                  const numRatio = typeof ratio === 'string' ? parseFloat(ratio) : ratio;
                  return numRatio <= 1 ? numRatio * 100 : numRatio;
                })()}
                total={100}
                showRate
                unit='%'
              />
            </InfoItem>
            <InfoItem>
              <InfoLabel>{t('CACHED')}</InfoLabel>
              <SimpleCircle
                theme="light"
                title={t('CACHED')}
                categories={[t('CACHED'), t('UFS_TOTAL')]}
                value={convertUnit(get(detail, 'status.cacheStates.cached', '-'))}
                total={convertUnit(get(detail, 'status.ufsTotal', '-'))}
                showRate
                unit='GiB'
              />
            </InfoItem>
          </InfoGrid>
        </Card>
      </CardWrapper>
      
      {/* 数据集拓扑图卡片 */}
      <CardWrapper>
        <Card sectionTitle={t('DATASET_TOPOLOGY')}>
          {renderTopologyGraph()}
        </Card>
      </CardWrapper>
      
      {/* 挂载信息卡片 */}
      {detail?.spec?.mounts && detail.spec.mounts.length > 0 && (
        <CardWrapper>
          <Card sectionTitle={t('MOUNT_INFORMATION')}>
            {detail.spec.mounts.map((mount: Mount, index: number) => (
              <MountCard key={index}>
                <MountHeader>
                  <MountTitle>{mount.name || `Mount ${index + 1}`}</MountTitle>
                </MountHeader>
                <MountDetails>
                  <MountItem>
                    <MountLabel>{t('MOUNT_POINT')}</MountLabel>
                    <MountValue>{mount.mountPoint}</MountValue>
                  </MountItem>
                  <MountItem>
                    <MountLabel>{t('PATH')}</MountLabel>
                    <MountValue>{mount.path || '-'}</MountValue>
                  </MountItem>
                  <MountItem>
                    <MountLabel>{t('READ_ONLY')}</MountLabel>
                    <MountValue>{mount.readOnly ? t('TRUE') : t('FALSE')}</MountValue>
                  </MountItem>
                  <MountItem>
                    <MountLabel>{t('SHARED')}</MountLabel>
                    <MountValue>{mount.shared ? t('TRUE') : t('FALSE')}</MountValue>
                  </MountItem>
                </MountDetails>
              </MountCard>
            ))}
          </Card>
        </CardWrapper>
      )}
    </>
  );
};

export default ResourceStatus; 