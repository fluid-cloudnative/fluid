import React from 'react';
import { Input, Switch, Row, Col } from '@kubed/components';
import { Trash } from '@kubed/icons';
import KVRecordInput, { validateKVPairs } from '../../../../../../../components/KVRecordInput';
import EncryptOptionsInput, { validateEncryptOptions } from './EncryptOptionsInput';
import { MountItemProps, EncryptOption } from '../types';
import {
  MountItem as StyledMountItem,
  RemoveButton,
  OptionsContainer,
  FormLabel,
  OptionalLabel,
  SwitchContainer,
  SwitchLabel,
} from '../styles';

declare const t: (key: string, options?: any) => string;

export const MountItem: React.FC<MountItemProps> = ({
  mount,
  index,
  canDelete,
  onUpdate,
  onRemove,
}) => {
  // 创建options验证函数
  const validateOptions = (options: Array<{ key: string; value: string }>) => {
    return validateKVPairs(options, {
      allowDuplicateKeys: false,
      allowEmptyKeys: false,
      allowEmptyValues: true
    });
  };

  return (
    <StyledMountItem>
      {canDelete && (
        <RemoveButton onClick={() => onRemove(index)}>
          <Trash size={25} />
        </RemoveButton>
      )}
      
      <Row gutter={[16, 16]}>
        <Col span={12}>
          <div style={{ marginBottom: '16px' }}>
            <FormLabel>
              {t('DATA_SOURCE')}
            </FormLabel>
            <Input
              value={mount.mountPoint}
              onChange={(e: React.ChangeEvent<HTMLInputElement>) => 
                onUpdate(index, 'mountPoint', e.target.value)
              }
              placeholder={t('DATA_SOURCE_PATH_PLACEHOLDER')}
            />
          </div>
        </Col>
        <Col span={4}>
          <div style={{ marginBottom: '16px' }}>
            <FormLabel>
              {t('NAME')}
            </FormLabel>
            <Input
              value={mount.name}
              onChange={(e: React.ChangeEvent<HTMLInputElement>) => 
                onUpdate(index, 'name', e.target.value)
              }
              placeholder="default"
            />
          </div>
        </Col>
        <Col span={8}>
          <div style={{ marginBottom: '16px' }}>
            <FormLabel>
              {t('MOUNT_PATH')}
              <OptionalLabel>
                {t("MOUNT_PATH_TIP") }/{mount.name || 'Name'}
              </OptionalLabel>
            </FormLabel>
            <Input
              value={mount.path}
              onChange={(e: React.ChangeEvent<HTMLInputElement>) => 
                onUpdate(index, 'path', e.target.value)
              }
              placeholder={`/${mount.name || 'Name'}`}
            />
          </div>
        </Col>
      </Row>

      <Row gutter={[16, 16]}>
        <Col span={4}>
          <SwitchContainer>
            <SwitchLabel>
              {t('READ_ONLY')}
            </SwitchLabel>
            <Switch
              checked={mount.readOnly}
              onChange={(checked) => onUpdate(index, 'readOnly', checked)}
            />
          </SwitchContainer>
        </Col>
        <Col span={4}>
          <SwitchContainer>
            <SwitchLabel>
              {t('SHARED')}
            </SwitchLabel>
            <Switch
              checked={mount.shared}
              onChange={(checked) => onUpdate(index, 'shared', checked)}
            />
          </SwitchContainer>
        </Col>
      </Row>

      <OptionsContainer>
        <div style={{ marginBottom: '16px' }}>
          <FormLabel>
            {t('OPTIONS')}
          </FormLabel>
          <KVRecordInput
            value={mount.options}
            onChange={(newOptions: Array<{ key: string; value: string }>) => {
              onUpdate(index, 'options', newOptions);
            }}
            validator={validateOptions}
            keyPlaceholder={t('OPTION_KEY')}
            valuePlaceholder={t('OPTION_VALUE')}
            addButtonText={t('ADD_OPTION')}
          />
        </div>

        <div style={{ marginBottom: '16px' }}>
          <FormLabel>
            {t('ENCRYPT_OPTIONS')}
            <OptionalLabel>
              {t("ENCRYPT_OPTIONS_tip")}
            </OptionalLabel>
          </FormLabel>
          <EncryptOptionsInput
            value={mount.encryptOptions || []}
            onChange={(newEncryptOptions: EncryptOption[]) => {
              onUpdate(index, 'encryptOptions', newEncryptOptions);
            }}
            validator={validateEncryptOptions}
          />
        </div>
      </OptionsContainer>
    </StyledMountItem>
  );
};
