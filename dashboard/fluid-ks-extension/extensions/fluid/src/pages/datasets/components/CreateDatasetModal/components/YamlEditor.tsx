import React, { useState, useEffect } from 'react';
import { Alert } from '@kubed/components';
import { CodeEditor } from '@kubed/code-editor';
import styled from 'styled-components';
import { DatasetFormData } from '../types';
import yaml from 'js-yaml';

declare const t: (key: string, options?: any) => string;

interface YamlEditorProps {
  formData: DatasetFormData;
  onDataChange: (data: DatasetFormData) => void;
  onValidationChange: (isValid: boolean) => void;
}

const EditorContainer = styled.div`
  padding: 24px;
  height: 600px;
  display: flex;
  flex-direction: column;
`;

const EditorHeader = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
`;

const EditorTitle = styled.h3`
  font-size: 16px;
  font-weight: 600;
  color: #242e42;
  margin: 0;
`;

const EditorWrapper = styled.div`
  flex: 1;
  border: 1px solid #e3e9ef;
  border-radius: 4px;
  overflow: hidden;
`;

const HiddenFileInput = styled.input`
  display: none;
`;

const YamlEditor: React.FC<YamlEditorProps> = ({
  formData,
  onDataChange,
  onValidationChange,
}) => {
  const [yamlContent, setYamlContent] = useState('');
  const [error, setError] = useState<string | null>(null);

  // 将表单数据转换为YAML
  const formDataToYaml = (data: DatasetFormData) => {
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
          name: data.runtimeName || data.name || '',
          namespace: data.namespace || 'default',
        },
      ],
    };

    const dataset = {
      apiVersion: 'data.fluid.io/v1alpha1',
      kind: 'Dataset',
      metadata: {
        name: data.name || '',
        namespace: data.namespace || 'default',
        labels: data.labels || {},
        ...(Object.keys(annotations).length > 0 && { annotations }),
      },
      spec: datasetSpec,
    };

    // 构建Runtime，优先使用完整的runtimeSpec，然后用UI字段覆盖
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
      kind: data.runtimeType || 'AlluxioRuntime',
      metadata: {
        name: data.runtimeName || data.name || '',
        namespace: data.namespace || 'default',
      },
      spec: runtimeSpec,
    };

    const resources = [dataset, runtime];

    // 如果启用了数据预热，添加DataLoad资源
    if (data.enableDataLoad && data.dataLoadConfig) {
      // 构建DataLoad spec，优先使用完整的dataLoadSpec，然后用UI字段覆盖
      const dataLoadSpec = {
        // 先使用用户在YAML中编辑的完整spec
        ...(data.dataLoadSpec || {}),
        // 固定字段：dataset绑定
        dataset: {
          name: data.name || '',
          namespace: data.namespace || 'default',
        },
        // UI字段优先级更高，覆盖YAML中的对应字段
        loadMetadata: data.dataLoadConfig.loadMetadata,
        target: data.dataLoadConfig.target || [],
        policy: data.dataLoadConfig.policy || 'Once',
        ...(data.dataLoadConfig.schedule && { schedule: data.dataLoadConfig.schedule }),
        ...(data.dataLoadConfig.ttlSecondsAfterFinished && { ttlSecondsAfterFinished: data.dataLoadConfig.ttlSecondsAfterFinished}),
      };

      const dataLoad: any = {
        apiVersion: 'data.fluid.io/v1alpha1',
        kind: 'DataLoad',
        metadata: {
          name: `${data.name}-dataload`,
          namespace: data.namespace || 'default',
          labels: {},
        },
        spec: dataLoadSpec,
      };
      resources.push(dataLoad);
    }

    return resources.map(resource => yaml.dump(resource)).join('---\n');
  };

  // 将YAML转换为表单数据
  const yamlToFormData = (yamlStr: string): DatasetFormData | null => {
    try {
      const documents = yamlStr.split('---').filter(doc => doc.trim());
      const resources = documents.map(doc => yaml.load(doc.trim()));
      
      const dataset = resources.find(r => r.kind === 'Dataset');
      const runtime = resources.find(r => r.kind?.endsWith('Runtime'));
      const dataLoad = resources.find(r => r.kind === 'DataLoad');

      if (!dataset) {
        throw new Error('Dataset resource not found');
      }

      // 提取完整的annotations，排除description字段
      const allAnnotations = { ...(dataset.metadata?.annotations || {}) };
      const description = allAnnotations['data.fluid.io/description'] || '';
      delete allAnnotations['data.fluid.io/description'];

      const formData: DatasetFormData = {
        name: dataset.metadata?.name || '',
        namespace: dataset.metadata?.namespace || 'default',
        description,
        labels: dataset.metadata?.labels || {},
        // 保存完整的annotations（除了description）
        annotations: Object.keys(allAnnotations).length > 0 ? allAnnotations : undefined,
        runtimeType: runtime?.kind || 'AlluxioRuntime',
        runtimeName: runtime?.metadata?.name || '',
        replicas: runtime?.spec?.replicas || 1,
        tieredStore: runtime?.spec?.tieredstore,
        // 保存完整的Runtime spec
        runtimeSpec: runtime?.spec ? { ...runtime.spec } : undefined,
        mounts: dataset.spec?.mounts || [],
        // 保存完整的Dataset spec
        datasetSpec: dataset.spec ? { ...dataset.spec } : undefined,
        enableDataLoad: !!dataLoad,
        dataLoadConfig: dataLoad ? {
          loadMetadata: dataLoad.spec?.loadMetadata || true,
          target: dataLoad.spec?.target || [],
          policy: dataLoad.spec?.policy || 'Once',
          schedule: dataLoad.spec?.schedule,
          ttlSecondsAfterFinished: dataLoad.spec?.ttlSecondsAfterFinished,
        } : undefined,
        // 保存完整的DataLoad spec
        dataLoadSpec: dataLoad?.spec ? { ...dataLoad.spec } : undefined,
      };

      return formData;
    } catch (err) {
      console.error('YAML parsing error:', err);
      return null;
    }
  };

  // 初始化YAML内容
  useEffect(() => {
    const yaml = formDataToYaml(formData);
    setYamlContent(yaml);
  }, []);

  // 处理YAML内容变化
  const handleYamlChange = (value: string) => {
    setYamlContent(value);
    setError(null);

    try {
      const newFormData = yamlToFormData(value);
      if (newFormData) {
        onDataChange(newFormData);
        onValidationChange(true);
      } else {
        onValidationChange(false);
        setError(t('YAML_PARSE_ERROR'));
      }
    } catch (err) {
      onValidationChange(false);
      setError((err as Error).message || t('YAML_PARSE_ERROR'));
    }
  };

  // // 下载YAML文件
  // const handleDownload = () => {
  //   const blob = new Blob([yamlContent], { type: 'text/yaml' });
  //   const url = URL.createObjectURL(blob);
  //   const a = document.createElement('a');
  //   a.href = url;
  //   a.download = `${formData.name || 'dataset'}.yaml`;
  //   document.body.appendChild(a);
  //   a.click();
  //   document.body.removeChild(a);
  //   URL.revokeObjectURL(url);
  // };

  // // 上传YAML文件
  // const handleUpload = (event: React.ChangeEvent<HTMLInputElement>) => {
  //   const file = event.target.files?.[0];
  //   if (file) {
  //     const reader = new FileReader();
  //     reader.onload = (e) => {
  //       const content = e.target?.result as string;
  //       handleYamlChange(content);
  //     };
  //     reader.readAsText(file);
  //   }
  //   // 清空input值，允许重复上传同一文件
  //   event.target.value = '';
  // };

  // // 重置为表单数据
  // const handleReset = () => {
  //   const yaml = formDataToYaml(formData);
  //   setYamlContent(yaml);
  //   setError(null);
  //   onValidationChange(true);
  // };

  return (
    <EditorContainer>
      <EditorHeader>
        <EditorTitle>{t('YAML_CONFIGURATION')}</EditorTitle>
      </EditorHeader>
      {error && (
        <Alert
          type="error"
          title={t('YAML_ERROR')}
          style={{ marginBottom: 16 }}
        >
          {error}  {'>_<'}
        </Alert>
      )}

      <EditorWrapper>
        <CodeEditor
          value={yamlContent}
          onChange={handleYamlChange}
        />
      </EditorWrapper>

      {/* <HiddenFileInput
        id="yaml-file-input"
        type="file"
        accept=".yaml,.yml"
        onChange={handleUpload}
      /> */}
    </EditorContainer>
  );
};

export default YamlEditor;
