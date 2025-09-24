import React, { useEffect, useState } from 'react';
import { Input, Select, InputNumber, Row, Col, Alert } from '@kubed/components';
import { StepComponentProps, RUNTIME_TYPE_OPTIONS, MEDIUM_TYPE_OPTIONS } from '../types';
import styled from 'styled-components';
import { Add, Trash } from '@kubed/icons';

declare const t: (key: string, options?: any) => string;

const StepContainer = styled.div`
  padding: 24px;
  min-height: 400px;
`;

const SectionTitle = styled.h4`
  font-size: 14px;
  font-weight: 600;
  color: #242e42;
  margin: 24px 0 16px 0;
  border-bottom: 1px solid #e3e9ef;
  padding-bottom: 8px;
`;

const TieredStoreItem = styled.div`
  border: 1px solid #e3e9ef;
  border-radius: 4px;
  padding: 16px;
  margin-bottom: 16px;
  background-color: #f9fbfd;
  position: relative;
`;

const RemoveButton = styled.button`
  position: absolute;
  top: 8px;
  right: 8px;
  background: none;
  border: none;
  color: #ca2621;
  cursor: pointer;
  padding: 4px;
  border-radius: 4px;
  
  &:hover {
    background-color: #fff2f2;
  }
`;

const AddTieredStoreLevelButton = styled.button`
  background: none;
  border: 1px dashed #d8dee5;
  color: #3385ff;
  padding: 12px 16px;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  
  &:hover {
    border-color: #3385ff;
    background-color: #f8faff;
  }
`;

const RuntimeStep: React.FC<StepComponentProps> = ({
  formData,
  onDataChange,
  onValidationChange,
}) => {
  const [formValues, setFormValues] = useState({
    runtimeType: 'AlluxioRuntime',
    replicas: 1,
  });

  // 初始化表单数据
  useEffect(() => {
    setFormValues({
      runtimeType: formData.runtimeType || 'AlluxioRuntime',
      replicas: formData.replicas || 1,
    });
  }, [formData]);

  // 初始化分层存储配置
  useEffect(() => {
    if (!formData.tieredStore) {
      onDataChange({
        tieredStore: {
          levels: [
            {
              level: 0,
              mediumtype: 'MEM',
              quota: '1Gi',
              path: '/dev/shm'
            },
          ],
        },
      });
    }
  }, []);

  // 表单值变化处理
  const handleFormChange = (field: string, value: any) => {
    const newValues = { ...formValues, [field]: value };
    setFormValues(newValues);

    // 只更新Runtime基本字段，保留完整的runtimeSpec
    onDataChange({
      runtimeType: newValues.runtimeType as any,
      runtimeName: formData.name, // 强制与数据集名称一致
      replicas: newValues.replicas,
      // 保留现有的tieredStore和runtimeSpec配置
    });

    // 验证表单
    const isValid = !!(newValues.runtimeType && newValues.replicas > 0);
    onValidationChange(isValid);
  };

  // 更新存储层配置
  const updateTieredStore = (levelIndex: number, field: string, value: any) => {
    const currentTieredStore = formData.tieredStore || { levels: [] };
    console.log("currenttieredstore:",currentTieredStore)
    const newLevels = [...currentTieredStore.levels];
    
    if (!newLevels[levelIndex]) {
      newLevels[levelIndex] = {
        level: levelIndex,
        mediumtype: 'MEM',
        quota: '2Gi',
      };
    }
    
    newLevels[levelIndex] = {
      ...newLevels[levelIndex],
      [field]: value,
    };

    onDataChange({
      tieredStore: {
        levels: newLevels,
      },
    });
  };

  // 添加存储层
  const addTieredStoreLevel = () => {
    const currentTieredStore = formData.tieredStore || { levels: [] };
    const newLevel = {
      level: currentTieredStore.levels.length,
      mediumtype: 'MEM',
      quota: '1Gi',
      path: '/dev/shm',
    };

    onDataChange({
      tieredStore: {
        levels: [...currentTieredStore.levels, newLevel],
      },
    });
  };

  // 删除存储层
  const removeTieredStoreLevel = (levelIndex: number) => {
    const currentTieredStore = formData.tieredStore || { levels: [] };
    const newLevels = currentTieredStore.levels.filter((_, index) => index !== levelIndex);
    
    // 重新编号
    newLevels.forEach((level, index) => {
      level.level = index;
    });

    onDataChange({
      tieredStore: {
        levels: newLevels,
      },
    });
  };

  const tieredStoreLevels = formData.tieredStore?.levels || [
    { level: 0, mediumtype: 'MEM', quota: '2Gi' },
  ];

  return (
    <StepContainer>
      <div>
        <Row gutter={[16, 0]}>
          <Alert
            type="info"
            title={t('RUNTIME_NAME_INFO')}
            style={{ marginBottom: 24 , marginRight: 100}}
          >
            {t('RUNTIME_NAME_INFO_DESC')}
          </Alert>
          <Col span={3}>
            <div style={{ marginBottom: '16px' }}>
              <label style={{ display: 'block', marginBottom: '8px', fontWeight: 600 }}>
                {t('RUNTIME_TYPE')} *
              </label>
              <Select
                placeholder={t('SELECT_RUNTIME_TYPE')}
                value={formValues.runtimeType}
                onChange={(value) => handleFormChange('runtimeType', value)}
                style={{ width: '90%' }}
              >
                {RUNTIME_TYPE_OPTIONS.map(option => (
                  <Select.Option key={option.value} value={option.value}>
                    {t(option.label)}
                  </Select.Option>
                ))}
              </Select>
            </div>
          </Col>
          <Col span={2}>
            <div style={{ marginBottom: '16px' }}>
              <label style={{ display: 'block', marginBottom: '8px', fontWeight: 600 }}>
                {t('REPLICAS')} *
              </label>
              <InputNumber
                min={1}
                max={100}
                value={formValues.replicas}
                onChange={(value) => handleFormChange('replicas', value)}
                placeholder={t('REPLICAS_PLACEHOLDER')}
                style={{ width: '100%' }}
              />
              {(!formValues.replicas || formValues.replicas < 1) && (
                <div style={{ color: '#ca2621', fontSize: '12px', marginTop: '4px' }}>
                  {t('REPLICAS_REQUIRED')}
                </div>
              )}
            </div>
          </Col>
        </Row>

        <SectionTitle>{t('TIERED_STORAGE')}</SectionTitle>
        
        {tieredStoreLevels.map((level, index) => (
          <TieredStoreItem key={index}>
            {tieredStoreLevels.length > 1 && (
              <RemoveButton onClick={() => removeTieredStoreLevel(index)}>
                <Trash size={25} />
              </RemoveButton>
            )}
            <Row gutter={[16, 0]} align="middle">
              <Col span={2}>
                <div style={{ marginBottom: '16px' }}>
                  <label style={{ display: 'block', marginBottom: '8px', fontWeight: 600 }}>
                    {t('LEVEL')}
                  </label>
                  <Input value={`Level ${level.level}`} disabled />
                </div>
              </Col>
              <Col span={2}>
                <div style={{ marginBottom: '16px' }}>
                  <label style={{ display: 'block', marginBottom: '8px', fontWeight: 600 }}>
                    {t('MEDIUM_TYPE')}
                  </label>
                  <Select
                    value={level.mediumtype}
                    onChange={(value) => updateTieredStore(index, 'mediumtype', value)}
                  >
                    {MEDIUM_TYPE_OPTIONS.map(option => (
                      <Select.Option key={option.value} value={option.value}>
                        {t(option.label)}
                      </Select.Option>
                    ))}
                  </Select>
                </div>
              </Col>
              <Col span={2}>
                <div style={{ marginBottom: '16px' }}>
                  <label style={{ display: 'block', marginBottom: '8px', fontWeight: 600 }}>
                    {t('QUOTA')}
                  </label>
                  <Input
                    value={level.quota}
                    onChange={(e: React.ChangeEvent<HTMLInputElement>) => updateTieredStore(index, 'quota', e.target.value)}
                    placeholder="2Gi"
                  />
                </div>
              </Col>
              <Col span={6}>
                <div style={{ marginBottom: '16px' }}>
                  <label style={{ display: 'block', marginBottom: '8px', fontWeight: 600 }}>
                    {t('PATH')}
                  </label>
                  <Input
                    value={level.path || ''}
                    onChange={(e: React.ChangeEvent<HTMLInputElement>) => updateTieredStore(index, 'path', e.target.value)}
                    placeholder={t('STORAGE_PATH_PLACEHOLDER')}
                  />
                </div>
              </Col>
            </Row>
          </TieredStoreItem>
        ))}

        <AddTieredStoreLevelButton onClick={addTieredStoreLevel}>
          <Add size={16} />
          {t('ADD_STORAGE_LEVEL')}
        </AddTieredStoreLevelButton>

        {/* <button
          type="button"
          onClick={addTieredStoreLevel}
          style={{
            background: 'none',
            border: '1px dashed #d8dee5',
            color: '#3385ff',
            padding: '8px 16px',
            borderRadius: '4px',
            cursor: 'pointer',
            fontSize: '14px',
            width: '100%',
          }}
        >
          + {t('ADD_STORAGE_LEVEL')}
        </button> */}
      </div>
    </StepContainer>
  );
};

export default RuntimeStep;
