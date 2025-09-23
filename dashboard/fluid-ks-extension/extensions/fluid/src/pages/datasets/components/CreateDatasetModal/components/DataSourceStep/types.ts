export interface SecretKeySelector {
  name: string;  // secret名称
  key: string;   // secret中的键
}

export interface EncryptOptionSource {
  secretKeyRef: SecretKeySelector;
}

export interface EncryptOption {
  name: string;  // 加密选项名称
  valueFrom?: EncryptOptionSource;
}

export interface Mount {
  mountPoint: string;
  name: string;
  path: string;
  readOnly: boolean;
  shared: boolean;
  options: Array<{ key: string; value: string }>;
  encryptOptions?: EncryptOption[];  // 新增字段
}

export interface MountItemProps {
  mount: Mount;
  index: number;
  canDelete: boolean;
  onUpdate: (index: number, field: keyof Mount, value: any) => void;
  onRemove: (index: number) => void;
}

export interface AddMountButtonProps {
  onAdd: () => void;
}
