/*
 * DataLoad detail page component
 */

import React, { useEffect, useMemo, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { Loading, Button } from '@kubed/components';
import { useCacheStore as useStore, yaml } from '@ks-console/shared';
import { DetailPagee } from '@ks-console/shared';
import { get } from 'lodash';
import { DownloadDuotone } from '@kubed/icons';

import { request } from '../../../utils/request';
import { EditYamlModal } from '@ks-console/shared';
import { handleResourceDelete } from '../../../utils/deleteResource';
import { createDetailTabs } from '../../../utils/detailTabs';

// 全局t函数声明
declare const t: (key: string, options?: any) => string;

// DataLoad类型定义
interface DataLoad {
  metadata: {
    name: string;
    namespace: string;
    creationTimestamp: string;
    uid: string;
    annotations?: Record<string, string>;
    labels?: Record<string, string>;
  };
  spec: {
    dataset: {
      name: string;
      namespace?: string;
    };
    target?: Array<{
      path: string;
      replicas?: number;
    }>;
    loadMetadata?: boolean;
    policy?: string;
  };
  status: {
    phase: string;
    duration: string;
    conditions: Array<{
      type: string;
      status: string;
      reason: string;
      message: string;
      lastProbeTime: string;
      lastTransitionTime: string;
    }>;
  };
}

const DataLoadDetail: React.FC = () => {
  const module = 'dataloads';
  const { cluster, namespace, name } = useParams<{ cluster: string; namespace: string; name: string }>();
  const navigate = useNavigate();
  const [dataload, setDataload] = useState<DataLoad | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<boolean>(false);
  const [editYamlConfig, setEditYamlConfig] = useState({
    editResource: null,
    visible: false,
    yaml: '',
    readOnly: true,
  });

  // 从URL参数获取集群信息
  const currentCluster = cluster || 'host';

  // 存储详情页数据到全局状态
  const [, setDetailProps] = useStore('DataLoadDetailProps', {
    module,
    detail: {},
    isLoading: false,
    isError: false,
  });

  // 列表页URL
  const listUrl = `/fluid/${currentCluster}/dataloads`;

  // 获取数据加载任务详情
  useEffect(() => {
    const fetchDataLoadDetail = async () => {
      console.log("fetchDataLoadDetail被调用了")
      try {
        setLoading(true);
        const response = await request(`/apis/data.fluid.io/v1alpha1/namespaces/${namespace}/dataloads/${name}`);
        if (!response.ok) {
          throw new Error(`Failed to fetch dataload: ${response.statusText}`);
        }
        const data = await response.json();
        setDataload(data);
        setError(false);
      } catch (error) {
        console.error('Failed to fetch dataload details:', error);
        setError(true);
      } finally {
        setLoading(false);
      }
    };

    if (namespace && name) {
      fetchDataLoadDetail();
    }
  }, [namespace, name, cluster]);

  // 更新全局状态
  useEffect(() => {
    setDetailProps({ 
      module, 
      detail: dataload as any, 
      isLoading: loading, 
      isError: error 
    });
  }, [dataload, loading, error]);

  // 定义标签页
  const tabs = useMemo(() => {
    return createDetailTabs(cluster || 'host', namespace!, name!, 'dataloads');
  }, [cluster, namespace, name]);

  // 编辑YAML
  const handleEditYaml = () => {
    if (!dataload) return;
    const yamlContent = yaml.getValue(dataload);
    setEditYamlConfig({
      editResource: dataload as any,
      visible: true,
      yaml: yamlContent,
      readOnly: true,
    });
  };

  // 删除资源
  const handleDelete = () => {
    if (!dataload) return;
    handleResourceDelete({
      resourceType: 'dataload',
      name: dataload.metadata.name,
      namespace: dataload.metadata.namespace,
      onSuccess: () => {
        navigate(listUrl);
      }
    });
  };

  // 操作按钮
  const actions = () => [
    {
      key: 'viewYaml',
      text: t('VIEW_YAML'),
      onClick: handleEditYaml,
    },
    {
      key: 'delete',
      text: t('DELETE'),
      render: () => (
        <Button color='error'
            onClick={handleDelete}>
              {t('DELETE')}
        </Button>
      )
    },
  ];

  // 属性信息
  const attrs = useMemo(() => {
    if (!dataload) return [];

    return [
      {
        label: t('CLUSTER'),
        value: currentCluster,
      },
      {
        label: t('PROJECT'),
        value: get(dataload, 'metadata.namespace', '-'),
      },
      {
        label: t('DATASET'),
        value: get(dataload, 'spec.dataset.name', '-'),
      },
      {
        label: t('STATUS'),
        value: get(dataload, 'status.phase', '-'),
      },
      {
        label: t('POLICY'),
        value: get(dataload, 'spec.policy', 'Once'),
      },
      {
        label: t('LOAD_METADATA'),
        value: get(dataload, 'spec.loadMetadata', false) ? t('TRUE') : t('FALSE'),
      },
      {
        label: t('DURATION'),
        value: get(dataload, 'status.duration', '-'),
      },
      {
        label: t('CREATION_TIME'),
        value: get(dataload, 'metadata.creationTimestamp', '-'),
      },
    ];
  }, [dataload, currentCluster]);

  return (
    <>
      {loading || error ? (
        <Loading className="page-loading" />
      ) : (
        <DetailPagee
          tabs={tabs}
          cardProps={{
            name: dataload?.metadata.name || '',
            params: { namespace, name },
            desc: get(dataload, 'metadata.annotations["kubesphere.io/description"]', ''),
            actions: actions(),
            attrs,
            breadcrumbs: {
              label: t('DATALOADS'),
              url: listUrl,
            },
            icon: <DownloadDuotone size={24}/>
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

export default DataLoadDetail;
