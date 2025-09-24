/*
 * Dataset detail page component
 */

import React, { useEffect, useMemo, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { Loading, Button } from '@kubed/components';
import { useCacheStore as useStore, yaml } from '@ks-console/shared';
import { DetailPagee } from '@ks-console/shared';
import { get } from 'lodash';
import { Book2Duotone } from '@kubed/icons';

import { request } from '../../../utils/request';
import { EditYamlModal } from '@ks-console/shared';
import { handleResourceDelete } from '../../../utils/deleteResource';
import { createDetailTabs } from '../../../utils/detailTabs';

// 全局t函数声明
declare const t: (key: string, options?: any) => string;

// 根据CRD定义更新Dataset类型
interface Dataset {
  metadata: {
    name: string;
    namespace: string;
    creationTimestamp: string;
    uid: string;
    labels?: Record<string, string>;
    annotations?: Record<string, string>;
  };
  spec: {
    mounts: Array<{
      mountPoint: string;
      name: string;
      options?: Record<string, string>;
      path?: string;
      readOnly?: boolean;
      shared?: boolean;
    }>;
    runtimes?: Array<{
      name: string;
      namespace: string;
      type: string;
      category?: string;
      masterReplicas?: number;
    }>;
  };
  status: {
    phase: string;
    conditions: Array<{
      type: string;
      status: string;
      reason: string;
      message: string;
      lastUpdateTime: string;
      lastTransitionTime: string;
    }>;
    cacheStates?: {
      cacheCapacity: string;
      cached: string;
      cachedPercentage: string;
      cacheHitRatio?: string;
    };
    ufsTotal?: string;
    fileNum?: string;
    hcfs?: {
      endpoint: string;
      underlayerFileSystemVersion?: string;
    };
  };
}

const DatasetDetail: React.FC = () => {
  const module = 'datasets';
  const { cluster, namespace, name } = useParams<{ cluster: string; namespace: string; name: string }>();
  const navigate = useNavigate();
  const [dataset, setDataset] = useState<Dataset | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<boolean>(false);
  const [editYamlConfig, setEditYamlConfig] = useState({
    editResource: null,
    visible: false,
    yaml: '',
    readOnly: true,
  });



  // 存储详情页数据到全局状态
  const [, setDetailProps] = useStore('DatasetDetailProps', {
    module,
    detail: {},
    isLoading: false,
    isError: false,
  });

  // 获取列表页URL
  const listUrl = useMemo(() => {
    const clusterName = cluster || 'host';
    return `/fluid/${clusterName}/datasets`;
  }, [cluster]);

  // 集群信息已从URL参数获取，无需额外同步

  // 获取数据集详情
  useEffect(() => {
    const fetchDatasetDetail = async () => {
      console.log("fetchDatasetDetail被调用了")
      try {
        setLoading(true);
        const response = await request(`/apis/data.fluid.io/v1alpha1/namespaces/${namespace}/datasets/${name}`);
        if (!response.ok) {
          throw new Error(`Failed to fetch dataset: ${response.statusText}`);
        }
        const data = await response.json();
        setDataset(data);
        setError(false);
      } catch (error) {
        console.error('Failed to fetch dataset details:', error);
        setError(true);
      } finally {
        setLoading(false);
      }
    };

    if (namespace && name) {
      fetchDatasetDetail();
    }
  }, [namespace, name, cluster]);

  // 更新全局状态
  useEffect(() => {
    setDetailProps({ 
      module, 
      detail: dataset as any, 
      isLoading: loading, 
      isError: error 
    });
  }, [dataset, loading, error]);

  // 定义标签页
  const tabs = useMemo(() => {
    return createDetailTabs(cluster || 'host', namespace!, name!, 'datasets');
  }, [cluster, namespace, name]);

  // 定义操作按钮
  const actions = () => {
    return [
      {
        key: 'viewYaml',
        text: t('VIEW_YAML'),
        onClick: () => {
          setEditYamlConfig({
            editResource: dataset as any,
            yaml: yaml.getValue(dataset as any),
            visible: true,
            readOnly: true,
          });
        },
      },
      {
        key: 'delete',
        render: () => (
          <Button
            color="error"
            onClick={() => {
              if (!dataset) return;

              handleResourceDelete({
                resourceType: 'dataset',
                name: dataset.metadata.name,
                namespace: dataset.metadata.namespace,
                onSuccess: () => {
                  // 删除成功后跳转回列表页
                  navigate(listUrl);
                }
              });
            }}
          >
            {t('DELETE')}
          </Button>
        ),
      },
    ];
  };

  // 定义属性
  const attrs = useMemo(() => {
    if (!dataset) return [];
    
    return [
      {
        label: t('STATUS'),
        value: get(dataset, 'status.phase', '-'),
      },
      {
        label: t('NAMESPACE'),
        value: dataset.metadata.namespace,
      },
      {
        label: t('UFS_TOTAL'),
        value: get(dataset, 'status.ufsTotal', '-'),
      },
      {
        label: t('CACHE_CAPACITY'),
        value: get(dataset, 'status.cacheStates.cacheCapacity', '-'),
      },
      {
        label: t('CACHED'),
        value: get(dataset, 'status.cacheStates.cached', '-'),
      },
      {
        label: t('CACHE_PERCENTAGE'),
        value: get(dataset, 'status.cacheStates.cachedPercentage', '-'),
      },
      {
        label: t('CACHE_HIT_RATIO'),
        value: get(dataset, 'status.cacheStates.cacheHitRatio', '-'),
      },
      {
        label: t('TOTAL_FILES'),
        value: get(dataset, 'status.fileNum', '-'),
      },
      {
        label: t('CREATION_TIME'),
        value: dataset.metadata.creationTimestamp,
      },
    ];
  }, [dataset]);

  return (
    <>
      {loading || error ? (
        <Loading className="page-loading" />
      ) : (
        <DetailPagee
          tabs={tabs}
          cardProps={{
            name: dataset?.metadata.name || '',
            params: { namespace, name },
            desc: get(dataset, 'metadata.annotations["kubesphere.io/description"]', ''),
            actions: actions(),
            attrs,
            breadcrumbs: {
              label: t('DATASETS'),
              url: listUrl,
            },
            icon: <Book2Duotone size={24}/>
          }}
        />
      )}
      {editYamlConfig.visible && (
        <EditYamlModal
          visible={editYamlConfig.visible}
          yaml={editYamlConfig.yaml}
          readOnly={editYamlConfig.readOnly}
          onCancel={() => setEditYamlConfig({ ...editYamlConfig, visible: false })}
        />
      )}
    </>
  );
};

export default DatasetDetail; 