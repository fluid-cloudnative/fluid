import React, { useState, useCallback } from 'react';
import { Button, Modal, Switch, notify, Steps, TabStep } from '@kubed/components';
import { Database, Cogwheel, RocketDuotone, DownloadDuotone, FolderDuotone, Book2Duotone } from '@kubed/icons';
import styled from 'styled-components';
import { getCurrentClusterFromUrl, request } from '../../../../utils/request';

const ModalContent = styled.div`
  max-height: calc(80vh - 120px);
  overflow-y: auto;
  position: relative;
  z-index: 1;

  .kubed-steps {
    position: relative;
    z-index: 1;
  }

  .kubed-tab-step {
    position: relative;
    z-index: 1;
  }

  input, textarea, select, .kubed-select {
    pointer-events: auto !important;
    position: relative;
    z-index: 2;
  }
`;

import DataLoadStep from '../../../datasets/components/CreateDatasetModal/components/DataLoadStep';
import { CreateDatasetModalProps, DatasetFormData, StepConfig } from '../../../datasets/components/CreateDatasetModal/types';
import DataLoadYamlEditor from './components/DataLoadYamlEditor';

declare const t: (key: string, options?: any) => string;

const STEPS: StepConfig[] = [
  {
    key: 'dataload',
    title: 'DATA_PRELOAD_CONFIGURATION',
    description: 'DATA_PRELOAD_CONFIG_DESC',
    component: DataLoadStep,
    icon: <DownloadDuotone size={24} />,
    optional: false,
  },
];

const CreateDataloadModal: React.FC<CreateDatasetModalProps> = ({
  visible,
  onCancel,
  onSuccess,
}) => {

    const [isYamlMode, setIsYamlMode] = useState(false);
    const [isCreating, setIsCreating] = useState(false);
    const [stepValidations, setStepValidations] = useState<Map<number, boolean>>(new Map());

    const [formData, setFormData] = useState<DatasetFormData>({
        // 项目开发初期我把数据集运行时数据加载这三个资源的字段都放入DatasetFormData中
        // 后面可以拆分，目前暂且用DatasetFormData来统一管理
        name: '',
        namespace: 'default',
        runtimeType: 'AlluxioRuntime',
        runtimeName: '',
        replicas: 1,
        enableDataLoad: true, // 独立创建DataLoad时默认启用
        dataLoadName: '',
        dataLoadNamespace: 'default',
        selectedDataset: '',
        selectedDatasetNamespace: 'default',
        dataLoadConfig: {
            loadMetadata: false,
            target: [ {path: "/", replicas: 1}],
            policy: 'Once',
            schedule: '',
            ttlSecondsAfterFinished: undefined,
        }
    });

    const handleDataChange = useCallback((data: Partial<DatasetFormData>) => {
        setFormData((prev: DatasetFormData) => {
          const newData = { ...prev, ...data };
          return newData;
        });
      }, []);

    const handleValidationChange = useCallback((stepIndex: number, isValid: boolean) => {
        setStepValidations(prev => {
          const newValidations = new Map(prev);
          newValidations.set(stepIndex, isValid);
          return newValidations;
        });
      }, []);

  // 重置表单
  const handleReset = () => {
    setStepValidations(new Map());
    setIsYamlMode(false);
    setFormData({
      name: '',
      namespace: 'default',
      runtimeType: 'AlluxioRuntime',
      runtimeName: '',
      replicas: 1,
      enableDataLoad: true, // 独立创建DataLoad时默认启用
      dataLoadName: '',
      dataLoadNamespace: 'default',
      selectedDataset: '',
      selectedDatasetNamespace: 'default',
      dataLoadConfig: {
      loadMetadata: false,
        target: [ {path: "/", replicas: 1}],
        policy: 'Once',
        schedule: '',
        ttlSecondsAfterFinished: undefined,
    }
    });
  };

    const handleClose = () => {
      handleReset();
      onCancel();
    };

    // 创建单个资源的API调用
    const createResource = async (resource: any, namespace: string) => {
        const clusterName = getCurrentClusterFromUrl();

        let url: string;
        if (resource.kind === 'DataLoad') {
            url = `/clusters/${clusterName}/apis/data.fluid.io/v1alpha1/namespaces/${namespace}/dataloads`;
        } else {
            throw new Error(`Unsupported resource kind: ${resource.kind}`);
        }

        const response = await fetch(url, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(resource),
        });

        if (!response.ok) {
            const errorText = await response.text();
            throw new Error(`Failed to create ${resource.kind}: ${response.status} ${response.statusText}\n${errorText}`);
        }

        return response.json();
    };

    // 将表单数据转换为DataLoad资源对象
    const formDataToDataLoadResource = (data: DatasetFormData) => {
        if (!data.dataLoadConfig) {
            throw new Error('DataLoad configuration is required');
        }

        const dataLoadSpec = {
            dataset: {
                name: data.selectedDataset || data.name,
                namespace: data.selectedDatasetNamespace || data.namespace,
            },
            loadMetadata: data.dataLoadConfig.loadMetadata,
            target: data.dataLoadConfig.target || [],
            policy: data.dataLoadConfig.policy || 'Once',
            ...(data.dataLoadConfig.schedule && { schedule: data.dataLoadConfig.schedule }),
            ...(data.dataLoadConfig?.ttlSecondsAfterFinished !== undefined && { ttlSecondsAfterFinished: data.dataLoadConfig?.ttlSecondsAfterFinished }),
        };

        const dataLoad = {
            apiVersion: 'data.fluid.io/v1alpha1',
            kind: 'DataLoad',
            metadata: {
                name: data.dataLoadName || `${data.selectedDataset || data.name}-dataload`,
                namespace: data.selectedDatasetNamespace,
            },
            spec: dataLoadSpec,
        };

        return dataLoad;
    };

    // 创建DataLoad
    const handleCreate = async () => {
        setIsCreating(true);
        try {
            console.log('Creating dataload with data:', formData);

            if (isYamlMode) {
                // YAML模式：从YamlEditor获取YAML内容并创建
                const yaml = await import('js-yaml');
                const resource = formDataToDataLoadResource(formData);
                const yamlContent = yaml.dump(resource);
                const documents = yamlContent.split('---').filter(doc => doc.trim());
                const resources = documents.map(doc => yaml.load(doc.trim()));

                for (const resource of resources) {
                    if (resource && typeof resource === 'object' && 'kind' in resource && 'metadata' in resource) {
                        console.log(`Creating ${resource.kind}:`, resource);
                        await createResource(resource, resource.metadata.namespace || formData.selectedDatasetNamespace);
                        console.log(`Successfully created ${resource.kind}: ${resource.metadata.name}`);
                    }
                }
            } else {
                // 表单模式：将表单数据转换为资源对象
                const resource = formDataToDataLoadResource(formData);
                console.log(`Creating ${resource.kind}:`, resource);
                await createResource(resource, formData.selectedDatasetNamespace as string);
                console.log(`Successfully created ${resource.kind}: ${resource.metadata.name}`);
            }

            notify.success(String(t('CREATE_DATALOAD_SUCCESS')));
            onSuccess?.();
            handleClose();
        } catch (error) {
            console.error('创建DataLoad失败:', error);
            notify.error(String(t('CREATE_DATALOAD_FAILED')) + ': ' + (error instanceof Error ? error.message : String(error)));
        } finally {
            setIsCreating(false);
        }
    };

    // 渲染步骤模式的底部按钮
    const renderStepsModalFooter = () => {
        const currentStepValid = stepValidations.get(0) !== false; // 默认为true，除非明确设置为false

        return (
            <>
                <Button variant="outline" onClick={handleClose}>
                    {t('CANCEL')}
                </Button>
                <Button
                    variant="filled"
                    color="success"
                    onClick={handleCreate}
                    loading={isCreating}
                    disabled={!currentStepValid}
                >
                    {t('CREATE')}
                </Button>
            </>
        );
    };

    return (
        <Modal
          title={
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', width: '100%', paddingRight: '40px' }}>
                <span>{t('CREATE_DATALOAD')}</span>
                <div style={{ display: 'flex', alignItems: 'center', gap: '24px' }}>
                    <a
                    href="https://github.com/fluid-cloudnative/fluid/blob/master/docs/en/dev/api_doc.md"
                    target="_blank"
                    rel="noopener noreferrer"
                    style={{ color: '#3385ff', textDecoration: 'none', fontSize: '14px' }}
                    >
                    {t("API_REFERENCE")}
                    </a>
                    <div style={{ display: 'flex', alignItems: 'center', gap: '8px', position: 'absolute', right: 70, top: 20 }}>
                    <span style={{ fontSize: '14px', color: '#79879c' }}>{t('YAML_MODE')}</span>
                    <Switch
                        checked={isYamlMode}
                        onChange={setIsYamlMode}
                    />
                    </div>
                </div>
                </div>
            }
            visible={visible}
            onCancel={handleClose}
            width={960}
            style={{ height: '80vh', maxHeight: '800px' }}
            footer={isYamlMode ? (
                <>
                <Button variant="outline" onClick={handleClose}>
                    {t('CANCEL')}
                </Button>
                <Button
                    variant="filled"
                    color="success"
                    onClick={handleCreate}
                    loading={isCreating}
                >
                    {t('CREATE')}
                </Button>
                </>
            ) : renderStepsModalFooter()}
            closable={true}
            maskClosable={false}
            >
            <ModalContent>
                {isYamlMode ? (
                <DataLoadYamlEditor
                    formData={formData}
                    onDataChange={handleDataChange}
                    onValidationChange={(isValid: boolean) => handleValidationChange(-1, isValid)}
                />
                ) : (
                <Steps active={0} variant="tab">
                    {STEPS.map((step, index) => (
                    <TabStep
                        key={step.key}
                        label={t(step.title)}
                        description={t(step.description)}
                        completedDescription={t('FINISHED')}
                        progressDescription={t('IN_PROGRESS')}
                        icon={step.icon}
                    >
                        <step.component
                            formData={formData}
                            onDataChange={handleDataChange}
                            onValidationChange={(isValid: boolean) => handleValidationChange(index, isValid)}
                            isIndependent={true}
                        />
                    </TabStep>
                    ))}
                </Steps>
                )}
            </ModalContent>
            </Modal>
    )
}

export default CreateDataloadModal;
