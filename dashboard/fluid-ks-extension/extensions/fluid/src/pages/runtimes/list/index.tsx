import React, { useState, useRef } from 'react';
import styled from 'styled-components';
import { get, debounce } from 'lodash';
import { Button, Card, Banner, Select, Empty, Modal, Input, notify } from '@kubed/components';
import { DataTable, TableRef, StatusIndicator } from '@ks-console/shared';
import { useNavigate, useParams } from 'react-router-dom';
import { RocketDuotone, SortLargeDuotone } from '@kubed/icons';
import { runtimeTypeList } from '../runtimeMap';
import { transformRequestParams } from '../../../utils';

import { getApiPath, getCurrentClusterFromUrl, requestPatch } from '../../../utils/request';
import { generateStatefulSetName } from '../../../utils/statefulSetUtils';
import { getStatusIndicatorType } from '../../../utils/getStatusIndicatorType';
import { useNamespaces } from '../../../utils/useNamespaces';
import { useWebSocketWatch } from '../../../utils/useWebSocketWatch';

// 声明全局 t 函数（国际化）
declare const t: (key: string) => string;

// 运行时类型存储键名
const RUNTIME_TYPE_STORAGE_KEY = 'fluid-runtime-type';

// 从 localStorage 获取上次选择的运行时类型
const getStoredRuntimeType = (): number => {
  try {
    const stored = localStorage.getItem(RUNTIME_TYPE_STORAGE_KEY);
    if (stored !== null) {
      const index = parseInt(stored, 10);
      // 验证索引是否在有效范围内
      if (index >= 0 && index < runtimeTypeList.length) {
        return index;
      }
    }
  } catch (error) {
    console.warn('Failed to read runtime type from localStorage:', error);
  }
  return 0; // 默认值：Alluxio
};

// 保存运行时类型到 localStorage
const saveRuntimeType = (index: number): void => {
  try {
    localStorage.setItem(RUNTIME_TYPE_STORAGE_KEY, index.toString());
  } catch (error) {
    console.warn('Failed to save runtime type to localStorage:', error);
  }
};

const StyledCard = styled(Card)`
  margin-bottom: 12px;
`;

const ToolbarWrapper = styled.div`
  display: flex;
  align-items: center;
  gap: 20px;
  margin-right: 20px;
`;

// Runtime 数据项接口
interface RuntimeItem {
  name: string;
  namespace: string;
  type: string; // runtime 类型，如 Alluxio、EFC 等
  masterReplicas: string | number;
  workerReplicas: string | number;
  creationTimestamp: string;
  masterPhase: string;
  workerPhase: string;
  fusePhase: string;
  raw: any; // 原始数据
  metadata: {
    name: string;
    namespace: string;
    uid: string;
  };
}

// Runtime 列表页组件
const RuntimeList: React.FC = () => {
  const navigate = useNavigate();
  const params = useParams<{ cluster: string }>();
  const [namespace, setNamespace] = useState<string>('');

  // 从URL参数获取集群信息
  const currentCluster = params.cluster || 'host';
  const [currentRuntimeType, setCurrentRuntimeType] = useState<number>(getStoredRuntimeType()); // 从localStorage获取上次选择的运行时类型
  // const [wsConnected, setWsConnected] = useState<boolean>(false);
  const tableRef = useRef<TableRef<any>>(null);

  // 扩缩容相关状态
  const [showScaleModal, setShowScaleModal] = useState<boolean>(false);
  const [selectedRuntime, setSelectedRuntime] = useState<RuntimeItem | null>(null);
  const [newReplicas, setNewReplicas] = useState<string>('');
  
  // 创建防抖的刷新函数，1000ms内最多执行一次
  const debouncedRefresh = debounce(() => {
    console.log("=== 执行防抖刷新 ===");
    if (tableRef.current) {
      tableRef.current.refetch();
    }
  }, 3000);

  // 自定义WebSocket实现来监控当前选中的运行时类型
  const { wsConnected } = useWebSocketWatch({
    namespace,
    resourcePlural: runtimeTypeList[currentRuntimeType].plural, // 根据当前选中的运行时类型动态获取 plural
    currentCluster,
    debouncedRefresh,
  });

  // 监听集群切换，刷新数据表格
  // const isFirstClusterEffect = useRef(true);
  // useEffect(() => {
  //   if (isFirstClusterEffect.current) {
  //     isFirstClusterEffect.current = false;
  //     return;
  //   }
  //   if (tableRef.current) {
  //     console.log('集群切换，刷新运行时数据表格:', currentCluster);
  //     debouncedRefresh();
  //   }
  // }, [currentCluster]);

  // 获取所有 namespace
  const { namespaces, isLoading, error } = useNamespaces(currentCluster)

  // 处理 namespace 变更
  const handleNamespaceChange = (value: string) => {
    setNamespace(value);
    debouncedRefresh();
  };

  // 处理 Runtime 类型切换
  const handleRuntimeTypeChange = (index: number) => {
    setCurrentRuntimeType(index);
    saveRuntimeType(index); // 保存到 localStorage
    debouncedRefresh();
  };

  // 点击名称跳转详情页
  const handleNameClick = (name: string, ns: string) => {
    const clusterName = params.cluster || 'host';
    const url = `/fluid/${clusterName}/${ns}/runtimes/${name}/resource-status`;
    navigate(url);
  };

  // 处理Master点击跳转
  const handleMasterClick = (runtimeName: string, namespace: string) => {
    const cluster = getCurrentClusterFromUrl();
    const currentRuntimeTypeMeta = runtimeTypeList[currentRuntimeType];
    const masterName = generateStatefulSetName(runtimeName, currentRuntimeTypeMeta.kind, 'master');
    const url = `/clusters/${cluster}/projects/${namespace}/statefulsets/${masterName}/resource-status`;
    console.log('Opening master in new window:', masterName, 'in namespace:', namespace, 'cluster:', cluster);
    window.open(url, '_blank');
  };

  // 处理Worker点击跳转
  const handleWorkerClick = (runtimeName: string, namespace: string) => {
    const cluster = getCurrentClusterFromUrl();
    const currentRuntimeTypeMeta = runtimeTypeList[currentRuntimeType];
    const workerName = generateStatefulSetName(runtimeName, currentRuntimeTypeMeta.kind, 'worker');
    const url = `/clusters/${cluster}/projects/${namespace}/statefulsets/${workerName}/resource-status`;
    console.log('Opening worker in new window:', workerName, 'in namespace:', namespace, 'cluster:', cluster);
    window.open(url, '_blank');
  };

  // 处理扩缩容按钮点击
  const handleScaleButtonClick = (runtime: RuntimeItem) => {
    setSelectedRuntime(runtime);
    setNewReplicas(runtime.workerReplicas.toString());
    setShowScaleModal(true);
  };

  // 处理扩缩容操作
  const handleScaleReplicas = async () => {
    if (!selectedRuntime || !newReplicas.trim()) {
      notify.error(t('REPLICAS_INPUT_REQUIRED'));
      return;
    }

    const replicas = parseInt(newReplicas, 10);
    if (isNaN(replicas) || replicas < 1) {
      notify.error(t('INVALID_REPLICAS_NUMBER'));
      return;
    }

    try {
      const currentRuntimeTypeMeta = runtimeTypeList[currentRuntimeType];
      const apiPath = `/apis/data.fluid.io/v1alpha1/namespaces/${selectedRuntime.namespace}/${currentRuntimeTypeMeta.plural}/${selectedRuntime.name}`;

      const patchBody = {
        spec: {
          replicas: replicas,
        },
      };

      await requestPatch(apiPath, patchBody);

      notify.success(`${t('SCALE_SUCCESS')}: ${selectedRuntime.name} -> ${replicas} replicas`);
      setShowScaleModal(false);
      setSelectedRuntime(null);
      setNewReplicas('');

      // 刷新表格数据
      if (tableRef.current) {
        tableRef.current.refetch();
      }
    } catch (error) {
      console.error('扩缩容失败:', error);
      const errorMessage = error instanceof Error ? error.message : String(error);
      notify.error(`${t('SCALE_FAILED')}: ${selectedRuntime.name} - ${errorMessage}`);
    }
  };

  // 格式化 Runtime 数据
  const formatRuntime = (item: any): RuntimeItem => {
    const typeMeta = runtimeTypeList[currentRuntimeType];
    return {
      name: get(item, 'metadata.name', ''),
      namespace: get(item, 'metadata.namespace', ''),
      type: typeMeta.displayName,
      masterReplicas: get(item, 'spec.master.replicas', get(item, 'spec.replicas', '-')),
      workerReplicas: get(item, 'spec.replicas', '-'),
      creationTimestamp: get(item, 'metadata.creationTimestamp', ''),
      masterPhase: get(item, 'status.masterPhase', '-'),
      workerPhase: get(item, 'status.workerPhase', '-'),
      fusePhase: get(item, 'status.fusePhase', '-'),
      raw: item,
      metadata: {
        name: get(item, 'metadata.name', ''),
        namespace: get(item, 'metadata.namespace', ''),
        uid: get(item, 'metadata.uid', `${get(item, 'metadata.namespace', '')}-${get(item, 'metadata.name', '')}-${typeMeta.kind}`),
      }
    };
  };

  // 表格列定义
  const columns = [
    {
      title: t('NAME'),
      field: 'name',
      width: '15%',
      searchable: true,
      render: (_: string, record: RuntimeItem) => (
        <a
          onClick={(e) => {
            e.preventDefault();
            handleNameClick(record.name, record.namespace);
          }}
          href="#"
        >
          {record.name}
        </a>
      ),
    },
    {
      title: t('NAMESPACE'),
      field: 'namespace',
      width: '15%',
      canHide: true,
    },
    {
      title: 'Replicas',
      field: 'workerReplicas',
      width: '15%',
      canHide: true,
      render: (value: string | number, record: RuntimeItem) => (
        <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
          <span>{value}</span>
          <Button
            size="sm"
            onClick={() => handleScaleButtonClick(record)}
            style={{ padding: '4px', minWidth: '24px', height: '24px' }}
            variant="text"
          >
            <SortLargeDuotone size={16}/>
          </Button>
        </div>
      ),
    },
    {
      title: 'MASTER PHASE',
      field: 'masterPhase',
      width: '13%',
      canHide: true,
      render: (value: string, record: RuntimeItem) => (
        <a
          onClick={(e) => {
            e.preventDefault();
            handleMasterClick(record.name, record.namespace);
          }}
          href="#"
          style={{ cursor: 'pointer' }}
        >
          <StatusIndicator type={getStatusIndicatorType(value)} motion={false}>
            {value || '-'}
          </StatusIndicator>

        </a>
      ),
    },
    {
      title: 'WORKER PHASE',
      field: 'workerPhase',
      width: '13%',
      canHide: true,
      render: (value: string, record: RuntimeItem) => (
        <a
          onClick={(e) => {
            e.preventDefault();
            handleWorkerClick(record.name, record.namespace);
          }}
          href="#"
          style={{ cursor: 'pointer' }}
        >
          <StatusIndicator type={getStatusIndicatorType(value)} motion={false}>
            {value || '-'}
          </StatusIndicator>
        </a>
      ),
    },
    {
      title: 'FUSE PHASE',
      field: 'fusePhase',
      width: '13%',
      canHide: true,
      render: (value: string) => (
        <StatusIndicator type={getStatusIndicatorType(value)} motion={false}>
            {value || '-'}
        </StatusIndicator>
      )
    },
    {
      title: t('CREATION_TIME'),
      field: 'creationTimestamp',
      width: '30%',
      canHide: true,
      sortable: true,
    },
  ] as any;

  // 获取 API 路径，添加集群前缀
  const basePath = namespace
    ? runtimeTypeList[currentRuntimeType].getApiPath(namespace)
    : runtimeTypeList[currentRuntimeType].getApiPath();
  const apiPath = getApiPath(basePath);

  return (
    <div>
      <Banner
        icon={<RocketDuotone/>}
        title={t('RUNTIMES')}
        description={t('RUNTIMES_DESC')}
        className="mb12"
      />

      {/* 连接状态指示器 */}
      <StatusIndicator type={wsConnected ? 'success' : 'warning'} motion={true}>
        {wsConnected ? t("WSCONNECTED_TIP") : t("WSDISCONNECTED_TIP")}
      </StatusIndicator>

      <StyledCard>
        {error ? (
          <Empty 
            icon="warning" 
            title={t('FETCH_ERROR_TITLE')} 
            description={error} 
            action={<Button onClick={debouncedRefresh}>{t('RETRY')}</Button>}
          />
        ) : (
          <DataTable
            ref={tableRef}
            rowKey="metadata.uid"
            tableName="runtimes-list"
            columns={columns}
            url={apiPath}
            format={formatRuntime}
            placeholder={t('SEARCH_BY_NAME')}
            transformRequestParams={transformRequestParams}
            simpleSearch={true}
            toolbarLeft={
              <ToolbarWrapper>
                <Select
                  value={namespace}
                  onChange={handleNamespaceChange}
                  placeholder={t('SELECT_NAMESPACE')}
                  style={{ width: 200 }}
                  disabled={isLoading}
                >
                  <Select.Option value="">{t('ALL_PROJECTS')}</Select.Option>
                  {namespaces.map(ns => (
                    <Select.Option key={ns} value={ns}>
                      {ns}
                    </Select.Option>
                  ))}
                </Select>
                <Select
                  value={currentRuntimeType}
                  onChange={handleRuntimeTypeChange}
                  placeholder={t('TYPE')}
                  style={{ width: 150 }}
                  disabled={isLoading}
                >
                  {runtimeTypeList.map((type, index) => (
                    <Select.Option key={type.kind} value={index}>
                      {type.displayName}
                    </Select.Option>
                  ))}
                </Select>
                
              </ToolbarWrapper>
            }
          />
        )}
      </StyledCard>

      {/* 扩缩容模态框 */}
      <Modal
        visible={showScaleModal}
        title={`${t('SCALE_RUNTIME_REPLICAS')}: ${selectedRuntime?.name || ''}`}
        onOk={handleScaleReplicas}
        onCancel={() => {
          setShowScaleModal(false);
          setSelectedRuntime(null);
          setNewReplicas('');
        }}
        width={500}
        closable={true}
        maskClosable={false}
      >
        <div style={{ padding: '24px 24px' }}>
          <Input
            label={t('WORKER_REPLICAS_COUNT')}
            placeholder={t('ENTER_NEW_REPLICAS')}
            value={newReplicas}
            onChange={(e: React.ChangeEvent<HTMLInputElement>) => setNewReplicas(e.target.value)}
            type="number"
            min={1}
          />
        </div>
      </Modal>
    </div>
  );
};

export default RuntimeList;