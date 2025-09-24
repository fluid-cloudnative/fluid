import React, { useState, useEffect } from 'react';
import { Alert } from '@kubed/components';
import { CodeEditor } from '@kubed/code-editor';
import styled from 'styled-components';
import { DatasetFormData } from '../../../../datasets/components/CreateDatasetModal/types';
import yaml from 'js-yaml';

declare const t: (key: string, options?: any) => string;

interface DataLoadYamlEditorProps {
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
  
  .kubed-code-editor {
    height: 100%;
  }
`;

const DataLoadYamlEditor: React.FC<DataLoadYamlEditorProps> = ({
  formData,
  onDataChange,
  onValidationChange,
}) => {
  const [yamlContent, setYamlContent] = useState<string>('');
  const [error, setError] = useState<string | null>(null);

  // 将表单数据转换为DataLoad YAML
  const formDataToYaml = (data: DatasetFormData) => {
    if (!data.dataLoadConfig) {
      // return '';
    }

    const dataLoadSpec = {
      dataset: {
        name: data.selectedDataset || data.name || '',
        namespace: data.selectedDatasetNamespace || data.namespace || 'default',
      },
      loadMetadata: data.dataLoadConfig?.loadMetadata,
      target: data.dataLoadConfig?.target || [],
      policy: data.dataLoadConfig?.policy || 'Once',
      ...(data.dataLoadConfig?.schedule && { schedule: data.dataLoadConfig.schedule }),
      ...(data.dataLoadConfig?.ttlSecondsAfterFinished !== undefined && { ttlSecondsAfterFinished: data.dataLoadConfig?.ttlSecondsAfterFinished }),
    };

    const dataLoad: any = {
      apiVersion: 'data.fluid.io/v1alpha1',
      kind: 'DataLoad',
      metadata: {
        name: data.dataLoadName || `${data.selectedDataset || data.name}-dataload`,
        namespace: data.selectedDatasetNamespace || data.namespace || 'default',
        labels: {},
      },
      spec: dataLoadSpec,
    };

    return yaml.dump(dataLoad);
  };

  // 将YAML转换为表单数据
  const yamlToFormData = (yamlStr: string): DatasetFormData | null => {
    try {
      const resource = yaml.load(yamlStr.trim());
      
      if (!resource || typeof resource !== 'object' || !('kind' in resource)) {
        throw new Error('Invalid YAML format');
      }

      if (resource.kind !== 'DataLoad') {
        throw new Error('YAML must contain a DataLoad resource');
      }

      const formData: DatasetFormData = {
        name: resource.metadata?.name || '',
        namespace: resource.metadata?.namespace || 'default',
        runtimeType: 'AlluxioRuntime',
        runtimeName: '',
        replicas: 1,
        dataLoadName: resource.metadata?.name || '',
        dataLoadNamespace: resource.metadata?.namespace || 'default',
        selectedDataset: resource.spec?.dataset?.name || '',
        selectedDatasetNamespace: resource.spec?.dataset?.namespace || 'default',
        enableDataLoad: true,
        dataLoadConfig: {
          loadMetadata: resource.spec?.loadMetadata || false,
          target: resource.spec?.target || [],
          policy: resource.spec?.policy || 'Once',
          schedule: resource.spec?.schedule,
          ttlSecondsAfterFinished: resource.spec?.ttlSecondsAfterFinished,
        },
        dataLoadSpec: resource.spec ? { ...resource.spec } : undefined,
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
          {error}
        </Alert>
      )}

      <EditorWrapper>
        <CodeEditor
          value={yamlContent}
          onChange={handleYamlChange}
        />
      </EditorWrapper>
    </EditorContainer>
  );
};

export default DataLoadYamlEditor;
