# Fluid文档

这个文件夹保存了[Fluid项目](https://github.com/fluid-cloudnative/fluid)文档。

- `en`: [英文文档](en/TOC.md)
- `zh`: [中文文档](zh/TOC.md)

## 文档版本

目前我们维护了如下版本的Fluid文档，每份都会对应相应分支。

| 分支名 | 版本描述 |
| :--- | :-- |
| `master` | 最新版 |
| `v0.1.0` | v0.1.0| 

## 生成PDF文档
目前我们提供了脚本以便用户自行生成PDF格式的文档。为了避免你配置生成环境，我们提供了Docker镜像，所以生成文档前，请确认你安装
了[Docker](https://www.docker.com/)。
1. 获取Docker镜像  
    `docker pull registry.cn-hangzhou.aliyuncs.com/docs-fluid/doc-build `
2. 创建容器  
    `docker run -it -v <your fluid/docs path>:/data/ fluid/doc-build:0.2.0`
3. 执行Makefile
    ```shell
   cd data
   make build
   make clean
    ```
