import React, { useState, useEffect, useRef } from 'react';
import styled from 'styled-components';
import { get, debounce } from 'lodash';
import { Button, Card, Banner, Select, Empty, Checkbox } from '@kubed/components';
import { DataTable, TableRef, StatusIndicator } from '@ks-console/shared';
import { useNavigate, useParams } from 'react-router-dom';
import { DownloadDuotone } from '@kubed/icons';
import { transformRequestParams } from '../../../utils';
import { deleteResource, handleBatchResourceDelete } from '../../../utils/deleteResource';

import { getApiPath, getWebSocketUrl, request, getCurrentClusterFromUrl } from '../../../utils/request';
import CreateDataloadModal from '../components/CreateDataloadModal';
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

// 根据CRD定义更新DataLoad类型
interface DataLoad {
  metadata: {
    name: string;
    namespace: string;
    creationTimestamp: string;
    uid: string;
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

// 格式化数据
const formatDataLoad = (item: Record<string, any>): DataLoad => {
  const dataload = {
    ...item,
    metadata: item.metadata || {},
    spec: item.spec || { 
      dataset: { name: '-' },
    },
    status: item.status || { 
      phase: '-', 
      duration: '-',
      conditions: []
    }
  };
  
  return dataload;
};

const DataLoadList: React.FC = () => {
  const [namespace, setNamespace] = useState<string>('');
  const [createModalVisible, setCreateModalVisible] = useState<boolean>(false);
  const [selectedDataLoads, setSelectedDataLoads] = useState<DataLoad[]>([]);
  const [isDeleting, setIsDeleting] = useState<boolean>(false);
  const [currentPageData, setCurrentPageData] = useState<DataLoad[]>([]);
  const [previousDataLength, setPreviousDataLength] = useState<number>(0);
  // const [wsConnected, setWsConnected] = useState<boolean>(false);
  const params: Record<string, any> = useParams();
  const navigate = useNavigate();
  const tableRef = useRef<TableRef<DataLoad>>(null);

  // 从URL参数获取集群信息
  const currentCluster = params.cluster || 'host';

  // 当命名空间变化时，清空选择状态和当前页面数据
  useEffect(() => {
    setSelectedDataLoads([]);
    // setCurrentPageData([]);
  }, [namespace]);

  // 监听数据变化，当数据加载任务数量发生变化时清空选择状态
  const handleDataChange = (newData: DataLoad[]) => {
    console.log("=== handleDataChange 被调用 ===");
    console.log("数据变化检测:", {
      previousLength: previousDataLength,
      newLength: newData?.length || 0,
      newData: newData
    });

    if (newData && previousDataLength > 0 && newData.length !== previousDataLength) {
      console.log("检测到数据加载任务数量变化，清空选择状态");
      setSelectedDataLoads([]);
    }

    setPreviousDataLength(newData?.length || 0);
    setCurrentPageData(newData || []);
  };

  // 创建防抖的刷新函数，1000ms内最多执行一次
  const debouncedRefresh = debounce(() => {
    console.log("=== 执行防抖刷新 ===");
    if (tableRef.current) {
      tableRef.current.refetch();
    }
  }, 1000);

  // 自定义WebSocket实现来替代DataTable的watchOptions
  const { wsConnected } = useWebSocketWatch({
    namespace,
    resourcePlural: 'dataloads',
    currentCluster,
    debouncedRefresh,
    onResourceDeleted: () => setSelectedDataLoads([]), // 当资源被删除时清空选择状态
  })

  // 用useNamespaces获取所有命名空间
  const { namespaces, isLoading, error, refetchNamespaces} = useNamespaces(currentCluster)

  // 处理命名空间变更
  const handleNamespaceChange = (value: string) => {
    setNamespace(value);
  };

  // 点击名称跳转到详情页的函数
  const handleNameClick = (name: string, ns: string) => {
    const currentCluster = getCurrentClusterFromUrl();
    const url = `/fluid/${currentCluster}/${ns}/dataloads/${name}/resource-status`;
    navigate(url);
  };
  
  // 创建数据加载任务按钮点击处理
  const handleCreateDataLoad = () => {
    setCreateModalVisible(true);
  };

  // 处理单个数据加载任务选择
  const handleSelectDataLoad = (dataload: DataLoad, checked: boolean) => {
    if (checked) {
      setSelectedDataLoads(prev => [...prev, dataload]);
    } else {
      const dataloadUid = get(dataload, 'metadata.uid', '');
      setSelectedDataLoads(prev => prev.filter(item => get(item, 'metadata.uid', '') !== dataloadUid));
    }
  };

  // 处理全选/取消全选
  const handleSelectAll = (checked: boolean) => {
    if (!checked) {
      // 取消全选
      setSelectedDataLoads([]);
    } else {
      // 全选：选择当前页面的所有数据加载任务
      setSelectedDataLoads([...currentPageData]);
    }
  };



  // 检查全选状态
  const isAllSelected = currentPageData.length > 0 && selectedDataLoads.length === currentPageData.length;
  const isIndeterminate = selectedDataLoads.length > 0 && selectedDataLoads.length < currentPageData.length;

  // 批量删除数据加载任务（使用通用删除函数）
  const handleBatchDelete = async () => {
    if (selectedDataLoads.length === 0) {
      return;
    }

    setIsDeleting(true);
    try {
      const resources = selectedDataLoads.map(dataload => ({
        name: get(dataload, 'metadata.name', ''),
        namespace: get(dataload, 'metadata.namespace', '')
      }));

      await handleBatchResourceDelete(resources, {
        resourceType: 'dataload',
        onSuccess: () => {
          setSelectedDataLoads([]);
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
      render: (_: any, record: DataLoad) => {
        const dataloadUid = get(record, 'metadata.uid', '');
        const isSelected = selectedDataLoads.some(item => get(item, 'metadata.uid', '') === dataloadUid);
        return (
          <Checkbox
            checked={isSelected}
            onChange={(e: React.ChangeEvent<HTMLInputElement>) => handleSelectDataLoad(record, e.target.checked)}
          />
        );
      },
    },
    {
      title: t('NAME'),
      field: 'metadata.name',
      width: '15%',
      searchable: true,
      render: (value: any, record: DataLoad) => (
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
      // render: (value: any, record: DataLoad) => <span>{get(record, 'metadata.namespace', '-')}</span>,
    },
    {
      title: t('DATASET'),
      field: 'spec.dataset.name',
      width: '15%',
      canHide: true,
      render: (value: any, record: DataLoad) => <span>{get(record, 'spec.dataset.name', '-')}</span>,
    },
    {
      title: t('STATUS'),
      field: 'status.phase',
      width: '10%',
      canHide: true,
      searchable: true,
      render: (value: any, record: DataLoad) => <span>{
        <StatusIndicator type={getStatusIndicatorType(value)} motion={false}>
            {value || '-'}
        </StatusIndicator>
      }</span>,
    },
    {
      title: t('POLICY'),
      field: 'spec.policy',
      width: '10%',
      canHide: true,
      render: (value: any, record: DataLoad) => <span>{get(record, 'spec.policy', 'Once')}</span>,
    },
    {
      title: t('LOAD_METADATA'),
      field: 'spec.loadMetadata',
      width: '10%',
      canHide: true,
      render: (value: any, record: DataLoad) => <span>{get(record, 'spec.loadMetadata', false) ? t('TRUE') : t('FALSE')}</span>,
    },
    {
      title: t('DURATION'),
      field: 'status.duration',
      width: '10%',
      canHide: true,
      sortable: true,
      render: (value: any, record: DataLoad) => <span>{get(record, 'status.duration', '-')}</span>,
    },
    {
      title: t('CREATION_TIME'),
      field: 'metadata.creationTimestamp',
      width: '15%',
      sortable: true,
      canHide: true,
      render: (value: any, record: DataLoad) => <span>{get(record, 'metadata.creationTimestamp', '-')}</span>,
    },
  ] as any;

  return (
    <div>
      <Banner
        icon={<DownloadDuotone/>}
        title={t('DATALOADS')}
        description={t('DATALOADS_DESC')}
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
            tableName="dataload-list"
            columns={columns}
            url={getApiPath(namespace ? `/kapis/data.fluid.io/v1alpha1/namespaces/${namespace}/dataloads` : '/kapis/data.fluid.io/v1alpha1/dataloads')}
            format={formatDataLoad}
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
                {selectedDataLoads.length > 0 && (
                  <Button
                    color="error"
                    onClick={handleBatchDelete}
                    loading={isDeleting}
                    style={{ marginRight: '8px' }}
                  >
                    {t('DELETE')} ({selectedDataLoads.length})
                  </Button>
                )}
                <Button onClick={handleCreateDataLoad}>
                  {t('CREATE_DATALOAD')}
                </Button>
              </div>
            }
          />
        )}
      </StyledCard>

      <CreateDataloadModal
        visible={createModalVisible}
        onCancel={() => setCreateModalVisible(false)}
        onSuccess={() => {
          setCreateModalVisible(false);
          // 刷新表格数据
          if (tableRef.current) {
            debouncedRefresh();
          }
        }}
      />


    </div>
  );
};

export default DataLoadList; 