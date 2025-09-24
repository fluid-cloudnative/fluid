import React, { useState, useCallback } from 'react';
import { Button, Modal, Switch, notify, Steps, TabStep } from '@kubed/components';
import { Database, Cogwheel, RocketDuotone, DownloadDuotone, FolderDuotone, Book2Duotone } from '@kubed/icons';
import styled from 'styled-components';

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
import BasicInfoStep from './components/BasicInfoStep';
import RuntimeStep from './components/RuntimeStep';
import DataSourceStep from './components/DataSourceStep/DataSourceStep';
import DataLoadStep from './components/DataLoadStep';
import YamlEditor from './components/YamlEditor';
import { CreateDatasetModalProps, DatasetFormData, StepConfig } from './types';
import { getCurrentCluster } from '../../../../utils/request';

declare const t: (key: string, options?: any) => string;

const STEPS: StepConfig[] = [
  {
    key: 'basic',
    title: 'BASIC_INFORMATION',
    description: 'BASIC_INFO_DESC',
    component: BasicInfoStep,
    icon: <Book2Duotone size={24} />,
  },
  {
    key: 'datasource',
    title: 'DATA_SOURCE_CONFIGURATION',
    description: 'DATA_SOURCE_CONFIG_DESC',
    component: DataSourceStep,
    icon: <FolderDuotone size={24} />,
  },
  {
    key: 'runtime',
    title: 'RUNTIME_CONFIGURATION',
    description: 'RUNTIME_CONFIG_DESC',
    component: RuntimeStep,
    icon: <RocketDuotone size={24} />,
  },
  {
    key: 'dataload',
    title: 'DATA_PRELOAD_CONFIGURATION',
    description: 'DATA_PRELOAD_CONFIG_DESC',
    component: DataLoadStep,
    icon: <DownloadDuotone size={24} />,
    optional: true,
  },
];

const CreateDatasetModal: React.FC<CreateDatasetModalProps> = ({
  visible,
  onCancel,
  onSuccess,
}) => {
  const [currentStep, setCurrentStep] = useState(0);
  const [completedSteps, setCompletedSteps] = useState<Set<number>>(new Set());
  const [stepValidations, setStepValidations] = useState<Record<number, boolean>>({});
  const [isYamlMode, setIsYamlMode] = useState(false);
  const [isCreating, setIsCreating] = useState(false);
  
  const [formData, setFormData] = useState<DatasetFormData>({
    name: '',
    namespace: 'default',
    runtimeType: 'AlluxioRuntime',
    runtimeName: '',
    replicas: 1,
    enableDataLoad: false,
  });

  // 更新表单数据
  const handleDataChange = useCallback((data: Partial<DatasetFormData>) => {
    setFormData(prev => {
      const newData = { ...prev, ...data };
      // 确保运行时名称与数据集名称一致
      if (data.name) {
        newData.runtimeName = data.name;
      }
      return newData;
    });
  }, []);

  // 更新步骤验证状态
  const handleValidationChange = useCallback((stepIndex: number, isValid: boolean) => {
    setStepValidations(prev => ({
      ...prev,
      [stepIndex]: isValid,
    }));
  }, []);

  // 渲染Modal footer
  const renderStepsModalFooter = () => {
    const isFirstStep = currentStep === 0;
    const isLastStep = currentStep === STEPS.length - 1;
    const currentStepValid = stepValidations[currentStep] !== false;
    const currentStepConfig = STEPS[currentStep];

    return (
      <>
        <Button variant="outline" onClick={handleClose}>
          {t('CANCEL')}
        </Button>
        {!isFirstStep && (
          <Button variant="outline" onClick={handlePrevious}>
            {t('PREVIOUS')}
          </Button>
        )}
        {currentStepConfig?.optional && !isLastStep &&(
          <Button variant="text" onClick={handleSkip}>
            {t('SKIP')}
          </Button>
        )}
        {!isLastStep && (
          <Button
            variant="filled"
            onClick={handleNext}
            disabled={!currentStepValid}
          >
            {t('NEXT')}
          </Button>
        )}
        {isLastStep && (
          <Button
            variant="filled"
            color="success"
            onClick={handleCreate}
            disabled={!currentStepValid}
            loading={isCreating}
          >
            {t('CREATE')}
          </Button>
        )}
      </>
    );
  };

  // 下一步
  const handleNext = () => {
    if (currentStep < STEPS.length - 1) {
      setCompletedSteps(prev => new Set([...prev, currentStep]));
      setCurrentStep(currentStep + 1);
    }
  };

  // 上一步
  const handlePrevious = () => {
    if (currentStep > 0) {
      setCompletedSteps(prev => {prev.delete(currentStep);return prev;});
      setCurrentStep(currentStep - 1);
    }
  };

  // 跳过当前步骤（仅对可选步骤有效）
  const handleSkip = () => {
    if (STEPS[currentStep].optional) {
      handleNext();
    }
  };

  // 获取集群名称（使用当前选择的集群）
  const getClusterName = () => {
    return getCurrentCluster();
  };

  // 创建单个资源的API调用
  const createResource = async (resource: any, namespace: string) => {
    const clusterName = getClusterName();

    let url: string;
    if (resource.kind === 'Dataset') {
      url = `/clusters/${clusterName}/apis/data.fluid.io/v1alpha1/namespaces/${namespace}/datasets`;
    } else if (resource.kind === 'DataLoad') {
      url = `/clusters/${clusterName}/apis/data.fluid.io/v1alpha1/namespaces/${namespace}/dataloads`;
    } else if (resource.kind.endsWith('Runtime')) {
      // 处理各种Runtime类型：AlluxioRuntime -> alluxioruntimes
      const runtimeType = resource.kind.toLowerCase() + 's';
      url = `/clusters/${clusterName}/apis/data.fluid.io/v1alpha1/namespaces/${namespace}/${runtimeType}`;
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

  // 将表单数据转换为资源对象
  const formDataToResources = (data: DatasetFormData) => {
    const resources: any[] = [];

    // 创建Dataset资源
    // 构建annotations，合并description和用户自定义的annotations
    const annotations: Record<string, string> = { ...(data.annotations || {}) };
    if (data.description) {
      annotations['data.fluid.io/description'] = data.description;
    }

    // 构建Dataset spec，优先使用完整的datasetSpec，然后用UI字段覆盖
    const datasetSpec = {
      // 先使用用户在YAML中编辑的完整spec
      ...(data.datasetSpec || {}),
      // UI字段优先级更高，覆盖YAML中的对应字段
      mounts: data.mounts || [],
      runtimes: [
        {
          name: data.runtimeName || data.name,
          namespace: data.namespace,
        },
      ],
    };

    const dataset = {
      apiVersion: 'data.fluid.io/v1alpha1',
      kind: 'Dataset',
      metadata: {
        name: data.name,
        namespace: data.namespace,
        labels: data.labels || {},
        ...(Object.keys(annotations).length > 0 && { annotations }),
      },
      spec: datasetSpec,
    };
    resources.push(dataset);

    // 创建Runtime资源
    // 构建Runtime spec，优先使用完整的runtimeSpec，然后用UI字段覆盖
    const defaultTieredStore = {
      levels: [
        {
          level: 0,
          mediumtype: 'MEM',
          quota: '2Gi',
        },
      ],
    };

    const runtimeSpec = {
      // 先使用用户在YAML中编辑的完整spec
      ...(data.runtimeSpec || {}),
      // UI字段优先级更高，覆盖YAML中的对应字段
      replicas: data.replicas || 1,
      tieredstore: data.tieredStore || defaultTieredStore,
    };

    const runtime = {
      apiVersion: 'data.fluid.io/v1alpha1',
      kind: data.runtimeType,
      metadata: {
        name: data.runtimeName || data.name,
        namespace: data.namespace,
      },
      spec: runtimeSpec,
    };
    resources.push(runtime);

    // 如果启用了数据预热，创建DataLoad资源
    if (data.enableDataLoad && data.dataLoadConfig) {
      // 构建DataLoad spec，优先使用完整的dataLoadSpec，然后用UI字段覆盖
      const dataLoadSpec = {
        // 先使用用户在YAML中编辑的完整spec
        ...(data.dataLoadSpec || {}),
        // 固定字段：dataset绑定
        dataset: {
          name: data.name,
          namespace: data.namespace,
        },
        // UI字段优先级更高，覆盖YAML中的对应字段
        loadMetadata: data.dataLoadConfig.loadMetadata,
        target: data.dataLoadConfig.target || [],
        policy: data.dataLoadConfig.policy || 'Once',
        ...(data.dataLoadConfig.schedule && { schedule: data.dataLoadConfig.schedule }),
      };

      const dataLoad = {
        apiVersion: 'data.fluid.io/v1alpha1',
        kind: 'DataLoad',
        metadata: {
          name: `${data.name}-dataload`,
          namespace: data.namespace,
        },
        spec: dataLoadSpec,
      };
      resources.push(dataLoad);
    }

    return resources;
  };

  // 从YAML内容创建资源
  const createFromYaml = async (yamlContent: string) => {
    const yaml = await import('js-yaml');
    const documents = yamlContent.split('---').filter(doc => doc.trim());
    const resources = documents.map(doc => yaml.load(doc.trim()));

    // 按顺序创建资源
    for (const resource of resources) {
      if (resource && typeof resource === 'object' && 'kind' in resource && 'metadata' in resource) {
        console.log(`Creating ${resource.kind}:`, resource);
        await createResource(resource, resource.metadata.namespace || formData.namespace);
        console.log(`Successfully created ${resource.kind}: ${resource.metadata.name}`);
      }
    }
  };

  // 创建数据集
  const handleCreate = async () => {
    setIsCreating(true);
    try {
      console.log('Creating dataset with data:', formData);

      if (isYamlMode) {
        // YAML模式：从YamlEditor获取YAML内容并创建
        // 注意：这里需要从YamlEditor组件获取当前的YAML内容
        // 由于YamlEditor已经验证了YAML并更新了formData，我们可以重新生成YAML
        const yaml = await import('js-yaml');
        const resources = formDataToResources(formData);
        const yamlContent = resources.map(resource => yaml.dump(resource)).join('---\n');
        await createFromYaml(yamlContent);
      } else {
        // 表单模式：将表单数据转换为资源对象
        const resources = formDataToResources(formData);

        // 按顺序创建资源：先创建Dataset，再创建Runtime，最后创建DataLoad
        for (const resource of resources) {
          console.log(`Creating ${resource.kind}:`, resource);
          await createResource(resource, formData.namespace);
          console.log(`Successfully created ${resource.kind}: ${resource.metadata.name}`);
        }
      }

      console.log('All resources created successfully');
      notify.success(t('CREATE_DATASET_SUCCESS') || 'Dataset created successfully');
      onSuccess?.();
      onCancel();
    } catch (error) {
      console.error('Failed to create dataset:', error);
      const errorMessage = error instanceof Error ? error.message : String(error);
      notify.error(`${t('CREATE_DATASET_FAILED') || 'Failed to create dataset'}: ${errorMessage}`);
    } finally {
      setIsCreating(false);
    }
  };

  // 重置表单
  const handleReset = () => {
    setCurrentStep(0);
    setCompletedSteps(new Set());
    setStepValidations({});
    setIsYamlMode(false);
    setFormData({
      name: '',
      namespace: 'default',
      runtimeType: 'AlluxioRuntime',
      runtimeName: '',
      replicas: 1,
      enableDataLoad: false,
    });
  };

  // 关闭Modal
  const handleClose = () => {
    handleReset();
    onCancel();
  };

  return (
    <Modal
      title={
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', width: '100%', paddingRight: '40px' }}>
          <span>{t('CREATE_DATASET')}</span>
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
          <YamlEditor
            formData={formData}
            onDataChange={handleDataChange}
            onValidationChange={(isValid) => handleValidationChange(-1, isValid)}
          />
        ) : (
          <Steps active={currentStep} variant="tab">
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
                />
              </TabStep>
            ))}
          </Steps>
        )}
      </ModalContent>
    </Modal>
  );
};

export default CreateDatasetModal;
