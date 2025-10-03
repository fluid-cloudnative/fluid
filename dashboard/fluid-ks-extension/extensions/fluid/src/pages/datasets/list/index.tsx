import React, { useState, useEffect, useRef } from 'react';
import styled from 'styled-components';
import { get, debounce } from 'lodash';
import { Button, Card, Banner, Select, Empty, Checkbox } from '@kubed/components';
import { DataTable, TableRef, StatusIndicator } from '@ks-console/shared';
import { useNavigate, useParams } from 'react-router-dom';
import { Book2Duotone } from '@kubed/icons';
import { transformRequestParams } from '../../../utils';
import CreateDatasetModal from '../components/CreateDatasetModal';
import { handleBatchResourceDelete } from '../../../utils/deleteResource';

import { getApiPath, getWebSocketUrl, request } from '../../../utils/request';
import { getStatusIndicatorType } from '../../../utils/getStatusIndicatorType';
import { useNamespaces } from '../../../utils/useNamespaces';
import { useWebSocketWatch } from '../../../utils/useWebSocketWatch';

// 全局t函数声明
declare const t: (key: string, options?: any) => string;

const StyledCard = styled(Card)`
  margin-bottom: 12px;
`;

const ToolbarWrapper = styled.div`
  display: flex;
  align-items: center;
  gap: 20px;
  margin-right: 20px;
`;

// 根据CRD定义更新Dataset类型
interface Dataset {
  metadata: {
    name: string;
    namespace: string;
    creationTimestamp: string;
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

// 格式化数据
const formatDataset = (item: Record<string, any>): Dataset => {
  const dataset = {
    ...item,
    metadata: item.metadata || {},
    spec: item.spec || { mounts: [] },
    status: item.status || { 
      phase: '-', 
      conditions: [],
      cacheStates: {
        cacheCapacity: '-',
        cached: '-',
        cachedPercentage: '-'
      }
    }
  };
  
  return dataset;
};

const DatasetList: React.FC = () => {
  const [namespace, setNamespace] = useState<string>('');
  const [createModalVisible, setCreateModalVisible] = useState<boolean>(false);
  const [selectedDatasets, setSelectedDatasets] = useState<Dataset[]>([]);
  const [isDeleting, setIsDeleting] = useState<boolean>(false);
  const [currentPageData, setCurrentPageData] = useState<Dataset[]>([]);
  const [previousDataLength, setPreviousDataLength] = useState<number>(0);
  // const [wsConnected, setWsConnected] = useState<boolean>(false);
  const params: Record<string, any> = useParams();
  const navigate = useNavigate();
  const tableRef = useRef<TableRef<Dataset>>(null);

  // 从URL参数获取集群信息
  const currentCluster = params.cluster || 'host';
  console.log(params,'params');
  
  // 用useNamespaces获取所有命名空间
  const { namespaces, isLoading, error, refetchNamespaces} = useNamespaces(currentCluster);

  // 当命名空间变化时，清空选择状态和当前页面数据
  useEffect(() => {
    setSelectedDatasets([]);
    // setCurrentPageData([]);
  }, [namespace]);

  // 监听数据变化，当数据集数量发生变化时清空选择状态
  const handleDataChange = (newData: Dataset[]) => {
    console.log("=== handleDataChange 被调用 ===");
    console.log("数据变化检测:", {
      previousLength: previousDataLength,
      newLength: newData?.length || 0,
      newData: newData
    });

    if (newData && previousDataLength > 0 && newData.length !== previousDataLength) {
      console.log("检测到数据集数量变化，清空选择状态");
      setSelectedDatasets([]);
    }

    setPreviousDataLength(newData?.length || 0);
    setCurrentPageData([...newData]);
  };



  // 创建防抖的刷新函数，1000ms内最多执行一次
  const debouncedRefresh = debounce(() => {
    console.log("=== 执行防抖刷新 ===");
    if (tableRef.current) {
      tableRef.current.refetch();
    }
  }, 1000);

  // 使用自定义WebSocket Hook 来替代DataTable的watchOptions
  const { wsConnected } = useWebSocketWatch({
    namespace,
    resourcePlural: 'datasets',
    currentCluster,
    debouncedRefresh,
    onResourceDeleted: () => setSelectedDatasets([]), // 当资源被删除时清空选择状态
  });

  // 处理命名空间变更
  const handleNamespaceChange = (value: string) => {
    setNamespace(value);
  };

  // 点击名称跳转到详情页的函数
  const handleNameClick = (name: string, ns: string) => {
    navigate(`/fluid/${currentCluster}/${ns}/datasets/${name}`);
  };
  
  // 创建数据集按钮点击处理
  const handleCreateDataset = () => {
    setCreateModalVisible(true);
  };

  // 创建数据集成功处理
  const handleCreateSuccess = () => {
    // 刷新表格数据
    // if (tableRef.current) {
    //   debouncedRefresh();
    // }
  };

  // 刷新表格数据
  const handleRefresh = () => {
    console.log("=== 手动刷新被调用 ===");
    debouncedRefresh();
  };

  // 处理单个数据集选择
  const handleSelectDataset = (dataset: Dataset, checked: boolean) => {
    if (checked) {
      setSelectedDatasets(prev => [...prev, dataset]);
    } else {
      const datasetUid = get(dataset, 'metadata.uid', '');
      setSelectedDatasets(prev => prev.filter(item => get(item, 'metadata.uid', '') !== datasetUid));
    }
  };

  // 处理全选/取消全选
  const handleSelectAll = (checked: boolean) => {
    if (!checked) {
      // 取消全选
      setSelectedDatasets([]);
    } else {
      setSelectedDatasets([...currentPageData]);
    }
  };

  // 检查全选状态
  const isAllSelected = currentPageData.length > 0 && selectedDatasets.length === currentPageData.length;
  const isIndeterminate = selectedDatasets.length > 0 && selectedDatasets.length < currentPageData.length;

  // 批量删除数据集（使用通用删除函数）
  const handleBatchDelete = async () => {
    if (selectedDatasets.length === 0) {
      return;
    }

    setIsDeleting(true);
    try {
      const resources = selectedDatasets.map(dataset => ({
        name: get(dataset, 'metadata.name', ''),
        namespace: get(dataset, 'metadata.namespace', '')
      }));

      await handleBatchResourceDelete(resources, {
        resourceType: 'dataset',
        onSuccess: () => {
          setSelectedDatasets([]);
          // 可以在这里添加刷新逻辑
        }
      });
    } finally {
      setIsDeleting(false);
    }
  };

  // 根据CRD定义完善表格列
  const columns = [
    {
      title: (
        <Checkbox
          checked={isAllSelected}
          indeterminate={isIndeterminate}
          onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
            handleSelectAll(e.target.checked);
          }}
        />
      ),
      field: 'selection',
      width: '50px',
      render: (_: any, record: Dataset) => {
        const datasetUid = get(record, 'metadata.uid', '');
        const isSelected = selectedDatasets.some(item => get(item, 'metadata.uid', '') === datasetUid);
        return (
          <Checkbox
            checked={isSelected}
            onChange={(e: React.ChangeEvent<HTMLInputElement>) => handleSelectDataset(record, e.target.checked)}
          />
        );
      },
    },
    {
      title: t('NAME'),
      field: 'metadata.name',
      width: '15%',
      searchable: true,
      render: (_: any, record: Dataset) => (
        <a
          onClick={(e) => {
            e.preventDefault();
            handleNameClick(get(record, 'metadata.name', ''), get(record, 'metadata.namespace', 'default'));
          }}
          href="#"
        >
          {get(record, 'metadata.name', '-')}
        </a>
      ),
    },
    {
      title: t('NAMESPACE'),
      field: 'metadata.namespace',
      width: '10%',
      canHide: true,
      render: (_: any, record: Dataset) => <span>{get(record, 'metadata.namespace', '-')}</span>,
    },
    {
      title: t('STATUS'),
      field: 'status.phase',
      width: '10%',
      canHide: true,
      searchable: true,
      sortable: true,
      render: (_: any, record: Dataset) => <span>{
        <StatusIndicator type={getStatusIndicatorType(get(record, 'status.phase', ''))} motion={false}>
          {get(record, 'status.phase', '-')}
        </StatusIndicator>
      }</span>,
      // render: (_: any, record: Dataset) => <span>{get(record, 'status.phase', '-')}</span>,
    },
    {
      title: t('DATA_SOURCE'),
      field: 'spec.mounts[0].mountPoint',
      width: '20%',
      canHide: true,
      searchable: true,
      render: (_: any, record: Dataset) => {
        const mountPoint = get(record, 'spec.mounts[0].mountPoint', '-');
        return <span>{mountPoint}</span>;
      },
    },
    {
      title: t('UFS_TOTAL'),
      field: 'status.ufsTotal',
      width: '10%',
      sortable: true,
      canHide: true,
      render: (_: any, record: Dataset) => <span>{get(record, 'status.ufsTotal', '-')}</span>,
    },
    {
      title: t('CACHE_CAPACITY'),
      field: 'status.cacheStates.cacheCapacity',
      width: '10%',
      sortable: true,
      canHide: true,
      render: (_: any, record: Dataset) => <span>{get(record, 'status.cacheStates.cacheCapacity', '-')}</span>,
    },
    {
      title: t('CACHED'),
      field: 'status.cacheStates.cached',
      width: '10%',
      sortable: true,
      canHide: true,
      render: (_: any, record: Dataset) => <span>{get(record, 'status.cacheStates.cached', '-')}</span>,
    },
    {
      title: t('CACHE_PERCENTAGE'),
      field: 'status.cacheStates.cachedPercentage',
      width: '10%',
      sortable: true,
      canHide: true,
      render: (_: any, record: Dataset) => <span>{get(record, 'status.cacheStates.cachedPercentage', '0%')}</span>,
    },
    {
      title: t('CREATION_TIME'),
      field: 'metadata.creationTimestamp',
      width: '10%',
      sortable: true,
      canHide: true,
      render: (_: any, record: Dataset) => <span>{get(record, 'metadata.creationTimestamp', '-')}</span>,
    },
  ] as any;

  return (
    <div>
      <Banner
        icon={<Book2Duotone />}
        title={t('DATASETS')}
        description={t('DATASET_DESC')}
        className="mb12"
      />
      <StatusIndicator type={wsConnected ? 'success' : 'warning'} motion={true}>
        {wsConnected ? t("WSCONNECTED_TIP") : t("WSDISCONNECTED_TIP")}
      </StatusIndicator>
      <StyledCard>
        {error ? (
          <Empty 
            icon="warning" 
            title={t('FETCH_ERROR_TITLE')} 
            description={error} 
            action={<Button onClick={handleRefresh}>{t('RETRY')}</Button>}
          />
        ) : (
          <DataTable
            ref={tableRef}
            rowKey="metadata.uid"
            tableName="dataset-list"
            columns={columns}
            url={getApiPath(namespace ? `/kapis/data.fluid.io/v1alpha1/namespaces/${namespace}/datasets` : '/kapis/data.fluid.io/v1alpha1/datasets')}
            format={formatDataset}
            placeholder={t('SEARCH_BY_NAME')}
            transformRequestParams={transformRequestParams}
            simpleSearch={true}
            onChangeData={handleDataChange}
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
              </ToolbarWrapper>
            }
            toolbarRight={
              <div style={{ display: 'flex', gap: '8px' }}>
                {selectedDatasets.length > 0 && (
                  <Button
                    color="error"
                    onClick={handleBatchDelete}
                    loading={isDeleting}
                    style={{ marginRight: '8px' }}
                  >
                    {t('DELETE')} ({selectedDatasets.length})
                  </Button>
                )}
                <Button onClick={handleCreateDataset}>
                  {t('CREATE_DATASET')}
                </Button>
              </div>
            }
          />
        )}
      </StyledCard>

      <CreateDatasetModal
        visible={createModalVisible}
        onCancel={() => setCreateModalVisible(false)}
        onSuccess={handleCreateSuccess}
      />
    </div>
  );
};

export default DatasetList; 