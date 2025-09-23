export interface DatasetFormData {
  // 基本信息
  name: string;
  namespace: string;
  description?: string;
  labels?: Record<string, string>;
  // 新增：完整的annotations支持，不仅仅是description
  annotations?: Record<string, string>;
  // 事实上这一部分对应的dataset的crd字段是：
  // metadata:
  //   name: string;
  //   namespace: string;
  //   annotations?: Record<string, string>; // 支持用户自定义的任意annotations
  //   labels?: Record<string, string>;

  // 数据集里的数据源配置
  mounts?: Array<{
    mountPoint: string;
    name: string;
    path?: string; // 挂载路径，如果不设置将是 /{Name}
    readOnly?: boolean;
    shared?: boolean;
    options?: Record<string, string>;
    encryptOptions?: Array<{
      name: string;
      valueFrom: {
        secretKeyRef: {
          name: string;
          key: string;
        };
      };
    }>;
  }>;
  // 这里对应dataset的crd字段mounts,fluid的api文档中提到This field can be empty because some runtimes don’t need to mount external storage (e.g. Vineyard).
  // mounts:
  // - mountPoint: local:///data-for-fluid
  //   name: my-local-data

  // 运行时配置
  runtimeType: RuntimeType;
  runtimeName: string; // 自动设置为数据集名称（fluid官网里的视频提到dataset和runtime名称要一样）
  replicas: number;
  // 新增：完整的Runtime spec支持，保存用户在YAML中编辑的所有字段
  runtimeSpec?: Record<string, any>;
  // 对应的是runtime的crd字段是：
  // spec:
  //   replicas: number;
  //   master?: {...};
  //   worker?: {...};
  //   fuse?: {...};
  //   properties?: {...};
  //   jvmOptions?: [...];
  //   ... 以及其他复杂的嵌套字段
  // 这些运行时的 replicas 字段均用于定义分布式缓存系统中工作节点的数量，直接影响数据缓存和处理的并行能力 (有必要向ui用户说明)
  
  // 存储配置
  tieredStore?: {
    levels: Array<{
      level: number;
      mediumtype: "MEM" | "SSD" | "HDD" | String;
      quota: string;
      path?: string;
      high?: Number;
      low?: Number;
      volumeType?: string;
    }>;
  };
  // 对应的runtime的字段应该是：
  // spec:
  //   tieredstore:
  //     levels:
  //         - mediumtype: MEM
  //           path: /dev/shm 
  //           quota: 1Gi
  //           high: "0.95"
  //           low: "0.7"
  //           volumeType: hostPath
  // path表示文件路径，用于定义该层级存储的路径,支持多个路径，多个路径用逗号分隔，如 “/mnt/cache1,/mnt/cache2”
  // 指定该存储层级使用的卷类型，可选值为 hostPath、emptyDir 和 volumeTemplate。若未明确设置，默认值为 hostPath
  
  // 数据预热配置
  enableDataLoad: boolean;
  dataLoadConfig?: {
    loadMetadata: boolean;
    target?: Array<{
      path: string;
      replicas?: number;
    }>;
    policy?: 'Once' | 'Cron' | 'OnEvent';
    schedule?: string;
    ttlSecondsAfterFinished?: number; // TTL seconds after finished
  };
  // 新增：完整的DataLoad spec支持，保存用户在YAML中编辑的所有字段
  dataLoadSpec?: Record<string, any>;
  // 对应dataload 的crd字段是
  // metadata:
  //   name:
  //   annotations?: Record<string, string>; // 支持用户自定义annotations
  //   labels?: Record<string, string>; // 支持用户自定义labels
  // spec:
  //   dataset:
  //     name:
  //     namespace: default
  //   loadMetadata?: boolean;
  //   target?: [...];
  //   policy?: string;
  //   schedule?: string;
  //   ... 以及其他高级字段
  // 由于是和dataset绑定，所以spec.dataset字段无需用户填写

  // 独立创建DataLoad时的额外字段
  dataLoadName?: string; // DataLoad的名称
  dataLoadNamespace?: string; // DataLoad所在的命名空间
  selectedDataset?: string; // 选择的数据集名称
  selectedDatasetNamespace?: string; // 选择的数据集所在的命名空间
  
  // 其他字段
  // Dataset高级设置 - 这些字段在API文档中定义但UI暂未支持，可通过YAML编辑
  // 以下字段将通过完整的Dataset spec支持，用户可在YAML模式下编辑：
  // owner?: User;                    // 数据集所有者
  // nodeAffinity?: CacheableNodeAffinity;  // 节点亲和性约束
  // tolerations?: Toleration[];      // 容忍度设置
  // accessModes?: PersistentVolumeAccessMode[];  // 访问模式
  // placement?: PlacementMode;       // 放置模式（多数据集单节点部署开关）
  // dataRestoreLocation?: DataRestoreLocation;  // 数据恢复位置
  // sharedOptions?: Record<string, string>;     // 共享选项
  // sharedEncryptOptions?: EncryptOption[];    // 共享加密选项

  // 新增：完整的Dataset spec支持，保存用户在YAML中编辑的所有字段
  datasetSpec?: Record<string, any>;
}

export type RuntimeType = 
  | 'AlluxioRuntime'
  | 'JindoRuntime' 
  | 'JuiceFSRuntime'
  | 'GooseFSRuntime'
  | 'EFCRuntime'
  | 'ThinRuntime'
  | 'VineyardRuntime';

export interface CreateDatasetModalProps {
  visible: boolean;
  onCancel: () => void;
  onSuccess?: () => void;
}

export interface StepComponentProps {
  formData: DatasetFormData;
  onDataChange: (data: Partial<DatasetFormData>) => void;
  onValidationChange: (isValid: boolean) => void;
  isIndependent?: boolean
}

export interface StepConfig {
  key: string;
  title: string;
  description: string;
  component: React.ComponentType<StepComponentProps>;
  optional?: boolean;
  icon?: React.ReactNode;
}

export const RUNTIME_TYPE_OPTIONS = [
  { label: 'Alluxio Runtime', value: 'AlluxioRuntime' },
  { label: 'Jindo Runtime', value: 'JindoRuntime' },
  { label: 'JuiceFS Runtime', value: 'JuiceFSRuntime' },
  { label: 'GooseFS Runtime', value: 'GooseFSRuntime' },
  { label: 'EFC Runtime', value: 'EFCRuntime' },
  { label: 'Thin Runtime', value: 'ThinRuntime' },
  { label: 'Vineyard Runtime', value: 'VineyardRuntime' },
];

export const MEDIUM_TYPE_OPTIONS = [
  { label: 'Memory', value: 'MEM' },
  { label: 'SSD', value: 'SSD' },
  { label: 'HDD', value: 'HDD' },
];

export const VOLUME_TYPE_OPTIONS = [
  { label: 'Host Path', value: 'hostPath' },
  { label: 'Empty Dir', value: 'emptyDir' },
  { label: 'PVC', value: 'persistentVolumeClaim' },
];
