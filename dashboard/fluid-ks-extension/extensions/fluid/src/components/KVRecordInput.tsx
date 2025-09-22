import React, { useState, useCallback } from 'react';
import { Input, Button } from '@kubed/components';
import { Trash, Add } from '@kubed/icons';
import styled from 'styled-components';


declare const t: (key: string, options?: any) => string;

// 键值对输入组件样式
const KVContainer = styled.div`
  margin-bottom: 16px;
`;

const KVItem = styled.div`
  display: flex;
  gap: 8px;
  margin-bottom: 8px;
  align-items: flex-start;
`;

const KVInputWrapper = styled.div`
  flex: 1;
  display: flex;
  gap: 8px;
`;

const KVInput = styled(Input)`
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

// 键值对验证结果接口
export interface KVValidationResult {
  valid: boolean;
  message?: string;
  duplicateIndexes?: number[]; // 重复键的索引数组
}

// 键值对输入组件接口
export interface KVRecordInputProps {
  value: Array<{ key: string; value: string }>;
  onChange: (value: Array<{ key: string; value: string }>) => void;
  validator?: (value: Array<{ key: string; value: string }>) => KVValidationResult;
  onError?: (error: KVValidationResult) => void;
  keyPlaceholder?: string;
  valuePlaceholder?: string;
  addButtonText?: string;
}

// 通用的键值对验证函数
export const validateKVPairs = (
  pairs: Array<{ key: string; value: string }>,
  options: {
    allowDuplicateKeys?: boolean;
    allowEmptyKeys?: boolean;
    allowEmptyValues?: boolean;
  } = {}
): KVValidationResult => {
  const {
    allowDuplicateKeys = false,
    allowEmptyKeys = false,
    allowEmptyValues = true
  } = options;

  // 检查重复键
  if (!allowDuplicateKeys) {
    const duplicateIndexes: number[] = [];
    const keyMap = new Map<string, number[]>();

    pairs.forEach((item, index) => {
      const key = item.key.trim();
      if (key !== '') {
        if (!keyMap.has(key)) {
          keyMap.set(key, []);
        }
        keyMap.get(key)!.push(index);
      }
    });

    keyMap.forEach((indexes) => {
      if (indexes.length > 1) {
        duplicateIndexes.push(...indexes);
      }
    });

    if (duplicateIndexes.length > 0) {
      return {
        valid: false,
        message: t('DUPLICATE_KEYS'),
        duplicateIndexes
      };
    }
  }

  // 检查空键
  if (!allowEmptyKeys) {
    const emptyKeyIndexes: number[] = [];
    pairs.forEach((item, index) => {
      if (item.value.trim() !== '' && item.key.trim() === '') {
        emptyKeyIndexes.push(index);
      }
    });

    if (emptyKeyIndexes.length > 0) {
      return {
        valid: false,
        message: t('EMPTY_KEY'),
        duplicateIndexes: emptyKeyIndexes
      };
    }
  }

  // 检查空值
  if (!allowEmptyValues) {
    const emptyValueIndexes: number[] = [];
    pairs.forEach((item, index) => {
      if (item.key.trim() !== '' && item.value.trim() === '') {
        emptyValueIndexes.push(index);
      }
    });

    if (emptyValueIndexes.length > 0) {
      return {
        valid: false,
        message: t('EMPTY_VALUE'),
        duplicateIndexes: emptyValueIndexes
      };
    }
  }

  return { valid: true };
};

const KVRecordInput: React.FC<KVRecordInputProps> = ({
  value = [],
  onChange,
  validator,
  onError,
  keyPlaceholder = t('KEY'),
  valuePlaceholder = t('VALUE'),
  addButtonText = t('ADD_LABEL')
}) => {
  const [validationError, setValidationError] = useState<KVValidationResult | null>(null);

  // 验证数据
  const validateData = useCallback((data: Array<{ key: string; value: string }>) => {
    if (validator) {
      const result = validator(data);
      setValidationError(result.valid ? null : result);
      if (onError) {
        onError(result);
      }
      return result;
    }
    setValidationError(null);
    return { valid: true };
  }, [validator, onError]);

  // 添加新项
  const addItem = useCallback(() => {
    const newValue = [...value, { key: '', value: '' }];
    onChange(newValue);
    validateData(newValue);
  }, [value, onChange, validateData]);

  // 删除项
  const removeItem = useCallback((index: number) => {
    const newValue = value.filter((_, i) => i !== index);
    onChange(newValue);
    validateData(newValue);
  }, [value, onChange, validateData]);

  // 更新项
  const updateItem = useCallback((index: number, field: 'key' | 'value', newValue: string) => {
    const updatedValue = [...value];
    updatedValue[index][field] = newValue;
    onChange(updatedValue);
    validateData(updatedValue);
  }, [value, onChange, validateData]);

  return (
    <KVContainer>
      {value.map((item, index) => {
        const isDuplicate = validationError?.duplicateIndexes?.includes(index);
        return (
          <KVItem key={`kv-${index}`}>
            <KVInputWrapper>
              <KVInput
                placeholder={keyPlaceholder}
                value={item.key}
                onChange={(e: React.ChangeEvent<HTMLInputElement>) =>
                  updateItem(index, 'key', e.target.value)
                }
                status={isDuplicate ? 'error' : undefined}
              />
              <KVInput
                placeholder={valuePlaceholder}
                value={item.value}
                onChange={(e: React.ChangeEvent<HTMLInputElement>) =>
                  updateItem(index, 'value', e.target.value)
                }
              />
            </KVInputWrapper>
            <DeleteButton
              type="button"
              onClick={() => removeItem(index)}
              title={t('REMOVE_LABEL')}
              // aria-label={`${t('REMOVE_LABEL')} ${index + 1}`}
            >
              <Trash size={16} />
            </DeleteButton>
          </KVItem>
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
        {addButtonText}
      </AddButton>
    </KVContainer>
    
  );
};

export default KVRecordInput;
