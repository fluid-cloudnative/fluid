// Fluid 支持的所有 Runtime 类型映射表
// 便于统一请求、类型判断和扩展

export interface RuntimeTypeMeta {
  kind: string;
  plural: string;
  apiVersion: string;
  displayName: string;
  // 获取 API 路径的方法，支持 namespace 级
  getApiPath: (namespace?: string) => string;
}

export const runtimeTypeList: RuntimeTypeMeta[] = [
  {
    kind: 'AlluxioRuntime',
    plural: 'alluxioruntimes',
    apiVersion: 'data.fluid.io/v1alpha1',
    displayName: 'Alluxio',
    getApiPath: (namespace?: string) =>
      namespace
        ? `/kapis/data.fluid.io/v1alpha1/namespaces/${namespace}/alluxioruntimes`
        : '/kapis/data.fluid.io/v1alpha1/alluxioruntimes',
  },
  {
    kind: 'EFCRuntime',
    plural: 'efcruntimes',
    apiVersion: 'data.fluid.io/v1alpha1',
    displayName: 'EFC',
    getApiPath: (namespace?: string) =>
      namespace
        ? `/kapis/data.fluid.io/v1alpha1/namespaces/${namespace}/efcruntimes`
        : '/kapis/data.fluid.io/v1alpha1/efcruntimes',
  },
  {
    kind: 'GooseFSRuntime',
    plural: 'goosefsruntimes',
    apiVersion: 'data.fluid.io/v1alpha1',
    displayName: 'GooseFS',
    getApiPath: (namespace?: string) =>
      namespace
        ? `/kapis/data.fluid.io/v1alpha1/namespaces/${namespace}/goosefsruntimes`
        : '/kapis/data.fluid.io/v1alpha1/goosefsruntimes',
  },
  {
    kind: 'JindoRuntime',
    plural: 'jindoruntimes',
    apiVersion: 'data.fluid.io/v1alpha1',
    displayName: 'Jindo',
    getApiPath: (namespace?: string) =>
      namespace
        ? `/kapis/data.fluid.io/v1alpha1/namespaces/${namespace}/jindoruntimes`
        : '/kapis/data.fluid.io/v1alpha1/jindoruntimes',
  },
  {
    kind: 'JuiceFSRuntime',
    plural: 'juicefsruntimes',
    apiVersion: 'data.fluid.io/v1alpha1',
    displayName: 'JuiceFS',
    getApiPath: (namespace?: string) =>
      namespace
        ? `/kapis/data.fluid.io/v1alpha1/namespaces/${namespace}/juicefsruntimes`
        : '/kapis/data.fluid.io/v1alpha1/juicefsruntimes',
  },
  {
    kind: 'ThinRuntime',
    plural: 'thinruntimes',
    apiVersion: 'data.fluid.io/v1alpha1',
    displayName: 'Thin',
    getApiPath: (namespace?: string) =>
      namespace
        ? `/kapis/data.fluid.io/v1alpha1/namespaces/${namespace}/thinruntimes`
        : '/kapis/data.fluid.io/v1alpha1/thinruntimes',
  },
  {
    kind: 'VineyardRuntime',
    plural: 'vineyardruntimes',
    apiVersion: 'data.fluid.io/v1alpha1',
    displayName: 'Vineyard',
    getApiPath: (namespace?: string) =>
      namespace
        ? `/kapis/data.fluid.io/v1alpha1/namespaces/${namespace}/vineyardruntimes`
        : '/kapis/data.fluid.io/v1alpha1/vineyardruntimes',
  },
]; 