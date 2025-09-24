import React, { useEffect, useState } from 'react';
import { StepComponentProps } from '../../types';
import { MountItem } from './components/MountItem';
import { AddMountButton } from './components/AddMountButton';
import { StepContainer } from './styles';
import { Mount } from './types';
import { initializeMountsFromFormData, convertMountsForSubmission, generateMountName } from './utils';


const DataSourceStep: React.FC<StepComponentProps> = ({
  formData,
  onDataChange,
  onValidationChange,
}) => {
  const [mounts, setMounts] = useState<Mount[]>(() =>
    initializeMountsFromFormData(formData.mounts)
  );

  // 验证挂载点是否填了
  useEffect(() => {
    const hasValidMount = mounts.some(mount => mount.mountPoint.trim() !== '');
    onValidationChange(hasValidMount);
  }, [mounts]); // 移除 onValidationChange 依赖，避免无限循环

  // 更新表单数据
  const updateFormData = (newMounts: Mount[]) => {
    const formattedMounts = convertMountsForSubmission(newMounts);
    onDataChange({ mounts: formattedMounts });
  };

  // 处理挂载点更新
  const handleMountUpdate = (index: number, field: keyof Mount, value: any) => {
    const newMounts = [...mounts];
    newMounts[index] = { ...newMounts[index], [field]: value };
    setMounts(newMounts);
    updateFormData(newMounts);
  };

  // 处理添加挂载点
  const handleAddMount = () => {
    const newMount: Mount = {
      mountPoint: '',
      path: '',
      readOnly: false,
      shared: true,
      options: [],
      encryptOptions: [],
      name: generateMountName(mounts),
    };
    const newMounts = [...mounts, newMount];
    setMounts(newMounts);
    updateFormData(newMounts);
  };

  // 处理删除挂载点
  const handleRemoveMount = (index: number) => {
    if (mounts.length > 1) {
      const newMounts = mounts.filter((_, i) => i !== index);
      setMounts(newMounts);
      updateFormData(newMounts);
    }
  };



  return (
    <StepContainer>
      {mounts.map((mount, index) => (
        <MountItem
          key={index}
          mount={mount}
          index={index}
          canDelete={mounts.length > 1}
          onUpdate={handleMountUpdate}
          onRemove={handleRemoveMount}
        />
      ))}

      <AddMountButton onAdd={handleAddMount} />
    </StepContainer>
  );
};

export default DataSourceStep;
