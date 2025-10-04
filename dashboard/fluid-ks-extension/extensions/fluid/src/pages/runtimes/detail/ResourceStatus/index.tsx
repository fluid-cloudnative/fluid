/*
 * Runtime Resource Status component
 */

import React from 'react';
import { useCacheStore as useStore, StatusIndicator } from '@ks-console/shared';
import { Card } from '@kubed/components';
import { DatabaseSealDuotone, StorageDuotone, AppstoreDuotone } from '@kubed/icons';
import { get } from 'lodash';
import { getCurrentClusterFromUrl } from '../../../../utils/request';
import { generateStatefulSetName } from '../../../../utils/statefulSetUtils';
import {
  CardWrapper,
  InfoGrid,
  InfoItem,
  InfoLabel,
  InfoValue,
  StatusCard,
  StatusHeader,
  StatusIcon,
  StatusTitle,
  StatusGrid,
  StatusItem,
  StatusValue,
  StatusLabel
} from '../../../shared/components/ResourceStatusStyles';
import { getStatusIndicatorType } from '../../../../utils/getStatusIndicatorType';

// 全局t函数声明
declare const t: (key: string, options?: any) => string;

const ResourceStatus = () => {
  const [props] = useStore('RuntimeDetailProps');
  const { detail, runtimeType } = props;

  if (!detail) {
    return <div>Loading runtime details...</div>;
  }

  // 获取运行时类型显示名称
  const runtimeTypeName = runtimeType?.displayName || 'Runtime';

  // 判断运行时是否有Master组件（ThinRuntime没有master）先留着
  const hasMaster = runtimeType?.kind !== 'ThinRuntime';

  // 获取整体状态 - 优先显示worker状态，worker如果ready那么master也一定ready
  // worker如果no ready，那么总体也就no ready
  const getOverallStatus = () => {
    // if (hasMaster) {
    //   return get(detail, 'status.masterPhase', '-');
    // }
    return get(detail, 'status.workerPhase', '-');
  };

  // 处理Master点击跳转
  const handleMasterClick = () => {
    const cluster = getCurrentClusterFromUrl();
    const namespace = get(detail, 'metadata.namespace', 'default');
    const runtimeName = get(detail, 'metadata.name', '');
    const masterName = generateStatefulSetName(runtimeName, runtimeType?.kind || '', 'master');
    const url = `/clusters/${cluster}/projects/${namespace}/statefulsets/${masterName}/resource-status`;
    console.log('Opening master in new window:', masterName, 'in namespace:', namespace, 'cluster:', cluster);
    window.open(url, '_blank');
  };

  // 处理Worker点击跳转
  const handleWorkerClick = () => {
    const cluster = getCurrentClusterFromUrl();
    const namespace = get(detail, 'metadata.namespace', 'default');
    const runtimeName = get(detail, 'metadata.name', '');
    const workerName = generateStatefulSetName(runtimeName, runtimeType?.kind || '', 'worker');
    const url = `/clusters/${cluster}/projects/${namespace}/statefulsets/${workerName}/resource-status`;
    console.log('Opening worker in new window:', workerName, 'in namespace:', namespace, 'cluster:', cluster);
    window.open(url, '_blank');
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
                    type={getStatusIndicatorType(getOverallStatus())}
                    motion={false}
                >
                  {getOverallStatus()}
                </StatusIndicator>
              </InfoValue>
            </InfoItem>
            <InfoItem>
              <InfoLabel>{t('TYPE')}</InfoLabel>
              <InfoValue>{runtimeTypeName}</InfoValue>
            </InfoItem>
            <InfoItem>
              <InfoLabel>{t('CREATION_TIME')}</InfoLabel>
              <InfoValue>{get(detail, 'metadata.creationTimestamp', '-')}</InfoValue>
            </InfoItem>
            <InfoItem>
              <InfoLabel>{t('SETUP_DURATION')}</InfoLabel>
              <InfoValue>{get(detail, 'status.setupDuration', '-')}</InfoValue>
            </InfoItem>
            <InfoItem>
              <InfoLabel>Worker {t('REPLICAS')}</InfoLabel>
              <InfoValue>{get(detail, 'spec.replicas', get(detail, 'spec.worker.replicas', '-'))}</InfoValue>
            </InfoItem>
            <InfoItem>
              <InfoLabel>{t('PROMETHEUS_MONITORING')}</InfoLabel>
              <InfoValue>{get(detail, 'spec.disablePrometheus', false) ? t('DISABLED') : t('ENABLED')}</InfoValue>
            </InfoItem>
          </InfoGrid>
        </Card>
      </CardWrapper>

      {/* Master 状态卡片 - 只在有Master组件时显示，可点击跳转 */}
      {hasMaster && (get(detail, 'spec.master') || get(detail, 'status.masterPhase')) && (
        <CardWrapper>
          <StatusCard style={{ cursor: 'pointer' }} onClick={handleMasterClick}>
            <StatusHeader>
              <StatusIcon>
                <DatabaseSealDuotone size={16} />
              </StatusIcon>
              <StatusTitle>Master ({t('CLICK_TO_VIEW_DETAILS')})</StatusTitle>
            </StatusHeader>
            <StatusGrid>
              <StatusItem>
                <StatusValue>
                  <StatusIndicator
                    type={getStatusIndicatorType(get(detail, 'status.masterPhase', ''))}
                    motion={false}
                  >
                    {get(detail, 'status.masterPhase', '-')}
                  </StatusIndicator>
                </StatusValue>
                <StatusLabel>{t('PHASE')}</StatusLabel>
              </StatusItem>
              <StatusItem>
                <StatusValue>
                    {get(detail, 'status.masterNumberReady', '0')}
                </StatusValue>
                <StatusLabel>{t('READY')}</StatusLabel>
              </StatusItem>
              <StatusItem>
                <StatusValue>{get(detail, 'status.desiredMasterNumberScheduled', '0')}</StatusValue>
                <StatusLabel>{t('DESIRED')}</StatusLabel>
              </StatusItem>
              <StatusItem>
                <StatusValue>{get(detail, 'status.currentMasterNumberScheduled', '0')}</StatusValue>
                <StatusLabel>{t('CURRENT')}</StatusLabel>
              </StatusItem>
              {get(detail, 'status.masterReason') && (
                <StatusItem>
                  <StatusValue>{get(detail, 'status.masterReason', '-')}</StatusValue>
                  <StatusLabel>{t('REASON')}</StatusLabel>
                </StatusItem>
              )}
            </StatusGrid>
          </StatusCard>
        </CardWrapper>
      )}

      {/* Worker 状态卡片 - 可点击跳转 */}
      {(get(detail, 'spec.worker') || get(detail, 'status.workerPhase') || get(detail, 'spec.replicas')) && (
        <CardWrapper>
          <StatusCard style={{ cursor: 'pointer' }} onClick={handleWorkerClick}>
            <StatusHeader>
              <StatusIcon>
                <StorageDuotone size={16} />
              </StatusIcon>
              <StatusTitle>Worker ({t('CLICK_TO_VIEW_DETAILS')})</StatusTitle>
            </StatusHeader>
            <StatusGrid>
              <StatusItem>
                <StatusValue>
                  <StatusIndicator
                    type={getStatusIndicatorType(get(detail, 'status.workerPhase', ''))}
                    motion={false}
                  >
                    {get(detail, 'status.workerPhase', '-')}
                  </StatusIndicator>
                </StatusValue>
                <StatusLabel>{t('PHASE')}</StatusLabel>
              </StatusItem>
              <StatusItem>
                <StatusValue>{get(detail, 'spec.replicas', get(detail, 'spec.worker.replicas', '0'))}</StatusValue>
                <StatusLabel>{t('REPLICAS')}</StatusLabel>
              </StatusItem>
              <StatusItem>
                <StatusValue>
                    {get(detail, 'status.workerNumberReady', '0')}
                </StatusValue>
                <StatusLabel>{t('READY')}</StatusLabel>
              </StatusItem>
              <StatusItem>
                <StatusValue>{get(detail, 'status.desiredWorkerNumberScheduled', '0')}</StatusValue>
                <StatusLabel>{t('DESIRED')}</StatusLabel>
              </StatusItem>
              <StatusItem>
                <StatusValue>{get(detail, 'status.currentWorkerNumberScheduled', '0')}</StatusValue>
                <StatusLabel>{t('CURRENT')}</StatusLabel>
              </StatusItem>
              {get(detail, 'status.workerNumberUnavailable', 0) > 0 && (
                <StatusItem>
                  <StatusValue>{get(detail, 'status.workerNumberUnavailable', '0')}</StatusValue>
                  <StatusLabel>{t('UNAVAILABLE')}</StatusLabel>
                </StatusItem>
              )}
              {get(detail, 'status.workerReason') && (
                <StatusItem>
                  <StatusValue>{get(detail, 'status.workerReason', '-')}</StatusValue>
                  <StatusLabel>{t('REASON')}</StatusLabel>
                </StatusItem>
              )}
            </StatusGrid>
          </StatusCard>
        </CardWrapper>
      )}

      {/* Fuse 状态卡片 */}
      {(get(detail, 'spec.fuse') || get(detail, 'status.fusePhase')) && (
        <CardWrapper>
          <StatusCard>
            <StatusHeader>
              <StatusIcon>
                <AppstoreDuotone size={16} />
              </StatusIcon>
              <StatusTitle>{runtimeType?.kind === 'VineyardRuntime' ? t('CLIENT_SOCKET') : 'Fuse'}</StatusTitle>
            </StatusHeader>
            <StatusGrid>
              <StatusItem>
                <StatusValue>
                  <StatusIndicator
                    type={getStatusIndicatorType(get(detail, 'status.fusePhase', ''))}
                    motion={false}
                  >
                    {get(detail, 'status.fusePhase', '-')}
                  </StatusIndicator>
                </StatusValue>
                <StatusLabel>{t('PHASE')}</StatusLabel>
              </StatusItem>
              <StatusItem>
                <StatusValue>
                    {get(detail, 'status.fuseNumberReady', '0')}
                </StatusValue>
                <StatusLabel>{t('READY')}</StatusLabel>
              </StatusItem>
              <StatusItem>
                <StatusValue>{get(detail, 'status.desiredFuseNumberScheduled', '0')}</StatusValue>
                <StatusLabel>{t('DESIRED')}</StatusLabel>
              </StatusItem>
              <StatusItem>
                <StatusValue>{get(detail, 'status.currentFuseNumberScheduled', '0')}</StatusValue>
                <StatusLabel>{t('CURRENT')}</StatusLabel>
              </StatusItem>
              {get(detail, 'status.fuseNumberUnavailable', 0) > 0 && (
                <StatusItem>
                  <StatusValue>{get(detail, 'status.fuseNumberUnavailable', '0')}</StatusValue>
                  <StatusLabel>{t('UNAVAILABLE')}</StatusLabel>
                </StatusItem>
              )}
              {get(detail, 'status.fuseNumberAvailable', 0) > 0 && (
                <StatusItem>
                  <StatusValue>{get(detail, 'status.fuseNumberAvailable', '0')}</StatusValue>
                  <StatusLabel>{t('AVAILABLE')}</StatusLabel>
                </StatusItem>
              )}
              {get(detail, 'status.fuseReason') && (
                <StatusItem>
                  <StatusValue>{get(detail, 'status.fuseReason', '-')}</StatusValue>
                  <StatusLabel>{t('REASON')}</StatusLabel>
                </StatusItem>
              )}
            </StatusGrid>
          </StatusCard>
        </CardWrapper>
      )}

      {/* 分层存储卡片 */}
      {get(detail, 'spec.tieredstore.levels') && (
        <CardWrapper>
          <Card sectionTitle={t('TIERED_STORAGE')}>
            <InfoGrid>
              {get(detail, 'spec.tieredstore.levels', []).map((level: any, index: number) => (
                <React.Fragment key={index}>
                  <InfoItem>
                    <InfoLabel>{t('LEVEL')} {level.level || index}</InfoLabel>
                    <InfoValue>-</InfoValue>
                  </InfoItem>
                  <InfoItem>
                    <InfoLabel>{t('MEDIUM_TYPE')}</InfoLabel>
                    <InfoValue>{level.mediumtype || '-'}</InfoValue>
                  </InfoItem>
                  <InfoItem>
                    <InfoLabel>{t('PATH')}</InfoLabel>
                    <InfoValue>{level.path || '-'}</InfoValue>
                  </InfoItem>
                  <InfoItem>
                    <InfoLabel>{t('QUOTA')}</InfoLabel>
                    <InfoValue>{level.quota || '-'}</InfoValue>
                  </InfoItem>
                  {level.volumeType && (
                    <InfoItem>
                      <InfoLabel>{t('VOLUME_TYPE')}</InfoLabel>
                      <InfoValue>{level.volumeType}</InfoValue>
                    </InfoItem>
                  )}
                  {level.high && (
                    <InfoItem>
                      <InfoLabel>High Watermark</InfoLabel>
                      <InfoValue>{level.high}</InfoValue>
                    </InfoItem>
                  )}
                  {level.low && (
                    <InfoItem>
                      <InfoLabel>Low Watermark</InfoLabel>
                      <InfoValue>{level.low}</InfoValue>
                    </InfoItem>
                  )}
                  {index < get(detail, 'spec.tieredstore.levels', []).length - 1 && (
                    <InfoItem style={{ gridColumn: '1 / -1', borderBottom: '1px solid #e3e9ef', margin: '8px 0' }}>
                      <InfoValue></InfoValue>
                    </InfoItem>
                  )}
                </React.Fragment>
              ))}
            </InfoGrid>
          </Card>
        </CardWrapper>
      )}

      {/* 缓存状态卡片 */}
      {get(detail, 'status.cacheStates') && (
        <CardWrapper>
          <Card sectionTitle={t('CACHE_STATUS')}>
            <InfoGrid>
              {Object.entries(get(detail, 'status.cacheStates', {})).map(([key, value]: [string, any]) => {
                // 缓存状态字段名映射
                const getCacheStateLabel = (fieldName: string) => {
                  const labelMap: Record<string, string> = {
                    'cached': t('CACHED'),
                    'cacheCapacity': t('CACHE_CAPACITY'),
                    'cacheHitRatio': t('CACHE_HIT_RATIO'),
                    'cachePercentage': t('CACHE_PERCENTAGE'),
                    'ufsTotal': t('UFS_TOTAL'),
                    'totalFiles': t('TOTAL_FILES'),
                    'cacheThroughputRatio': t('cacheThroughputRatio'),
                  };
                  return labelMap[fieldName] || fieldName;
                };

                return (
                  <InfoItem key={key}>
                    <InfoLabel>{getCacheStateLabel(key)}</InfoLabel>
                    <InfoValue>{typeof value === 'object' ? JSON.stringify(value) : String(value)}</InfoValue>
                  </InfoItem>
                );
              })}
            </InfoGrid>
          </Card>
        </CardWrapper>
      )}
    </>
  );
};

export default ResourceStatus;
