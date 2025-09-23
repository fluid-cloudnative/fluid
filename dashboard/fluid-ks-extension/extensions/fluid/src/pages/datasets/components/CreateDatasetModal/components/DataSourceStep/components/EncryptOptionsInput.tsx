import React, { useState } from 'react';
import { Input, Button } from '@kubed/components';
import { Trash, Add } from '@kubed/icons';
import styled from 'styled-components';
import { EncryptOption } from '../types';

declare const t: (key: string, options?: any) => string;

// 样式定义
const EncryptContainer = styled.div`
  margin-bottom: 16px;
`;

const EncryptItem = styled.div`
  display: flex;
  gap: 8px;
  margin-bottom: 8px;
  align-items: flex-start;
`;

const EncryptInputWrapper = styled.div`
  flex: 1;
  display: flex;
  gap: 8px;
`;

const EncryptInput = styled(Input)`
  flex: 1;
`;

const DeleteButton = styled.button`
  background: none;
  border: none;
  color: #ca2621;
  cursor: pointer;
  padding: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
  min-width: 32px;
  height: 32px;

  &:hover {
    background-color: #ffeaea;
  }
`;

const AddButton = styled(Button)`
  border: 1px dashed #d8dee5;
  color: #3385ff;
  background: none;

  &:hover {
    border-color: #3385ff;
    background-color: #f8faff;
    color: #3385ff;
  }
`;

const ValidationError = styled.div`
  color: #ca2621;
  font-size: 12px;
  margin-top: 8px;
  padding: 8px 12px;
  background-color: #ffeaea;
  border: 1px solid #ffcdd2;
  border-radius: 4px;
  display: flex;
  align-items: center;
  gap: 8px;
`;

// 验证结果接口
export interface EncryptValidationResult {
  valid: boolean;
  message?: string;
  errorIndexes?: number[];
}

// 组件接口
export interface EncryptOptionsInputProps {
  value: EncryptOption[];
  onChange: (value: EncryptOption[]) => void;
  validator?: (value: EncryptOption[]) => EncryptValidationResult;
  onError?: (error: EncryptValidationResult) => void;
}

// 验证函数
export const validateEncryptOptions = (
  options: EncryptOption[]
): EncryptValidationResult => {
  const errorIndexes: number[] = [];
  const nameMap = new Map<string, number[]>();

  // 检查每个选项
  options.forEach((option, index) => {
    // 检查name是否为空
    if (!option.name.trim()) {
      errorIndexes.push(index);
      return;
    }

    // 检查valueFrom是否完整
    if (option.valueFrom) {
      if (!option.valueFrom.secretKeyRef.name.trim() || !option.valueFrom.secretKeyRef.key.trim()) {
        errorIndexes.push(index);
        return;
      }
    }

    // 收集name用于重复检查
    const name = option.name.trim();
    if (!nameMap.has(name)) {
      nameMap.set(name, []);
    }
    nameMap.get(name)!.push(index);
  });

  // 检查重复的name
  nameMap.forEach((indexes) => {
    if (indexes.length > 1) {
      errorIndexes.push(...indexes);
    }
  });

  if (errorIndexes.length > 0) {
    return {
      valid: false,
      message: t('ENCRYPT_OPTIONS_VALIDATION_ERROR'),
      errorIndexes
    };
  }

  return { valid: true };
};

const EncryptOptionsInput: React.FC<EncryptOptionsInputProps> = ({
  value = [],
  onChange,
  validator,
  onError,
}) => {
  const [validationError, setValidationError] = useState<EncryptValidationResult | null>(null);

  // 验证数据
  function validateData(data: EncryptOption[]) {
    const result = validator ? validator(data) : validateEncryptOptions(data);
    setValidationError(result.valid ? null : result);
    if (onError) {
      onError(result);
    }
    return result;
  }

  // 添加新项
  function addItem() {
    const newValue = [...value, {
      name: '',
      valueFrom: {
        secretKeyRef: {
          name: '',
          key: ''
        }
      }
    }];
    onChange(newValue);
    validateData(newValue);
  }

  // 删除项
  function removeItem(index: number) {
    const newValue = value.filter((_, i) => i !== index);
    onChange(newValue);
    validateData(newValue);
  }

  // 更新项
  function updateItem(index: number, field: string, newValue: string) {
    const updatedValue = [...value];
    const option = { ...updatedValue[index] };

    if (field === 'name') {
      option.name = newValue;
    } else if (field === 'secretName') {
      if (!option.valueFrom) {
        option.valueFrom = { secretKeyRef: { name: '', key: '' } };
      }
      option.valueFrom.secretKeyRef.name = newValue;
    } else if (field === 'secretKey') {
      if (!option.valueFrom) {
        option.valueFrom = { secretKeyRef: { name: '', key: '' } };
      }
      option.valueFrom.secretKeyRef.key = newValue;
    }

    updatedValue[index] = option;
    onChange(updatedValue);
    validateData(updatedValue);
  }

  return (
    <EncryptContainer>
      {value.map((item, index) => {
        const hasError = validationError?.errorIndexes?.includes(index);
        return (
          <EncryptItem key={`encrypt-${index}`}>
            <EncryptInputWrapper>
              <EncryptInput
                placeholder={t('ENCRYPT_OPTION_NAME')}
                value={item.name}
                onChange={(e: React.ChangeEvent<HTMLInputElement>) =>
                  updateItem(index, 'name', e.target.value)
                }
                status={hasError ? 'error' : undefined}
              />
              <EncryptInput
                placeholder={t('SECRET_NAME')}
                value={item.valueFrom?.secretKeyRef.name || ''}
                onChange={(e: React.ChangeEvent<HTMLInputElement>) =>
                  updateItem(index, 'secretName', e.target.value)
                }
                status={hasError ? 'error' : undefined}
              />
              <EncryptInput
                placeholder={t('SECRET_KEY')}
                value={item.valueFrom?.secretKeyRef.key || ''}
                onChange={(e: React.ChangeEvent<HTMLInputElement>) =>
                  updateItem(index, 'secretKey', e.target.value)
                }
                status={hasError ? 'error' : undefined}
              />
            </EncryptInputWrapper>
            <DeleteButton
              type="button"
              onClick={() => removeItem(index)}
              title={t('REMOVE_LABEL')}
            >
              <Trash size={16} />
            </DeleteButton>
          </EncryptItem>
        );
      })}

      {/* 显示验证错误 */}
      {validationError && !validationError.valid && (
        <ValidationError>
          <span>⚠️</span>
          <span>{validationError.message}</span>
        </ValidationError>
      )}

      <AddButton
        type="button"
        onClick={addItem}
      >
        <Add size={16} style={{ marginRight: '4px' }} />
        {t('ADD_ENCRYPT_OPTION')}
      </AddButton>
    </EncryptContainer>
  );
};

export default EncryptOptionsInput;
