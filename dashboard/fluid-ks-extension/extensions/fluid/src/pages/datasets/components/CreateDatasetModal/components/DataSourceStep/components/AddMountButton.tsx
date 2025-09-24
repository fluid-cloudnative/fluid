import React from 'react';
import { Add } from '@kubed/icons';
import { AddMountButton as StyledAddMountButton } from '../styles';
import { AddMountButtonProps } from '../types';

declare const t: (key: string, options?: any) => string;

export const AddMountButton: React.FC<AddMountButtonProps> = ({ onAdd }) => {
  return (
    <StyledAddMountButton onClick={onAdd}>
      <Add size={18} />
      {t('ADD_MOUNT_POINT')}
    </StyledAddMountButton>
  );
};
