/*
 * Runtime detail page component
 */

import React, { useEffect, useMemo, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { Loading } from '@kubed/components';
import { useCacheStore as useStore, yaml } from '@ks-console/shared';
import { DetailPagee } from '@ks-console/shared';
import { get } from 'lodash';
import { RocketDuotone } from '@kubed/icons';

import { request } from '../../../utils/request';
import { EditYamlModal } from '@ks-console/shared';
import { runtimeTypeList, RuntimeTypeMeta } from '../runtimeMap';
import { createDetailTabs } from '../../../utils/detailTabs';

// 全局t函数声明
declare const t: (key: string, options?: any) => string;

// Runtime 接口定义
interface Runtime {
  apiVersion: string;
  kind: string;
  metadata: {
    name: string;
    namespace: string;
    uid: string;
    creationTimestamp: string;
    labels?: Record<string, string>;
    annotations?: Record<string, string>;
  };
  spec: any;
  status: any;
}

// 运行时类型检测结果
interface RuntimeTypeInfo {
  runtime: Runtime;
  typeMeta: RuntimeTypeMeta;
}

const RuntimeDetail: React.FC = () => {
  const module = 'runtimes';
  const { cluster, namespace, name } = useParams<{ cluster: string; namespace: string; name: string }>();
  const navigate = useNavigate();
  const [runtimeInfo, setRuntimeInfo] = useState<RuntimeTypeInfo | null>(null);
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
  const [, setDetailProps] = useStore('RuntimeDetailProps', {
    module,
    detail: {},
    isLoading: false,
    isError: false,
  });

  // 列表页URL
  const listUrl = `/fluid/${currentCluster}/runtimes`;

  // 动态检测运行时类型并获取详情
  useEffect(() => {
    const fetchRuntimeDetail = async () => {
      console.log("fetchRuntimeDetail被调用了")
      try {
        setLoading(true);
        
        // 尝试不同类型的运行时API
        for (const typeMeta of runtimeTypeList) {
          try {
            const apiPath = `/apis/data.fluid.io/v1alpha1/namespaces/${namespace}/${typeMeta.plural}/${name}`;
            const response = await request(apiPath);
            
            if (response.ok) {
              const data = await response.json();
              setRuntimeInfo({ runtime: data, typeMeta });
              setError(false);
              return; // 找到了，退出循环
            }
          } catch (err) {
            // 继续尝试下一个类型
            console.log(`Failed to fetch ${typeMeta.kind}:`, err);
          }
        }
        
        // 如果所有类型都尝试失败了
        throw new Error('Runtime not found in any supported type');
        
      } catch (error) {
        console.error('Failed to fetch runtime details:', error);
        setError(true);
      } finally {
        setLoading(false);
      }
    };

    if (namespace && name) {
      fetchRuntimeDetail();
    }
  }, [namespace, name, cluster]);

  // 更新全局状态
  useEffect(() => {
    setDetailProps({
      module,
      detail: runtimeInfo?.runtime as any,
      isLoading: loading,
      isError: error
    });
    // 将runtimeType单独存储
    if (runtimeInfo?.typeMeta) {
      (setDetailProps as any)((prev: any) => ({ ...prev, runtimeType: runtimeInfo.typeMeta }));
    }
  }, [runtimeInfo, loading, error]);

  // 定义标签页
  const tabs = useMemo(() => {
    return createDetailTabs(cluster || 'host', namespace!, name!, 'runtimes');
  }, [cluster, namespace, name]);

  // 定义操作按钮（不包含删除按钮）
  const actions = () => {
    return [
      {
        key: 'viewYaml',
        text: t('VIEW_YAML'),
        onClick: () => {
          setEditYamlConfig({
            editResource: runtimeInfo?.runtime as any,
            yaml: yaml.getValue(runtimeInfo?.runtime as any),
            visible: true,
            readOnly: true,
          });
        },
      },
    ];
  };

  // 定义属性
  const attrs = useMemo(() => {
    if (!runtimeInfo) return [];

    const { runtime, typeMeta } = runtimeInfo;

    return [
      {
        label: t('STATUS'),
        value: get(runtime, 'status.workerPhase', '-'),
      },
      {
        label: t('CLUSTER'),
        value: currentCluster,
      },
      {
        label: t('PROJECT'),
        value: get(runtime, 'metadata.namespace', '-'),
      },
      {
        label: t('TYPE'),
        value: typeMeta.displayName,
      },
      {
        label: t('API_VERSION'),
        value: get(runtime, 'apiVersion', '-'),
      },
      {
        label: t('RESOURCE_VERSION'),
        value: get(runtime, 'metadata.resourceVersion', '-'),
      },
      {
        label: t('GENERATION'),
        value: get(runtime, 'metadata.generation', '-'),
      },
      {
        label: t('WORKER_REPLICAS'),
        value: get(runtime, 'spec.replicas', get(runtime, 'spec.worker.replicas', '-')),
      },
      {
        label: t('CREATION_TIME'),
        value: get(runtime, 'metadata.creationTimestamp', '-'),
      },
    ];
  }, [runtimeInfo, currentCluster]);

  return (
    <>
      {loading || error ? (
        <Loading className="page-loading" />
      ) : (
        <DetailPagee
          tabs={tabs}
          cardProps={{
            name: runtimeInfo?.runtime?.metadata.name || '',
            params: { namespace, name },
            desc: get(runtimeInfo?.runtime, 'metadata.annotations["kubesphere.io/description"]', ''),
            actions: actions(),
            attrs,
            breadcrumbs: {
              label: t('RUNTIMES'),
              url: listUrl,
            },
            icon: <RocketDuotone size={24}/>
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

export default RuntimeDetail;
