# AI科研专项基金项目需求书

## 面向大模型推理的云原生状态感知调度与弹性资源编排系统

---

## 1. 需求题目

**基于Fluid的大语言模型(LLM)推理服务状态感知调度与存算分离架构的智能资源编排系统研究**

---

## 2. 业务背景

### 2.1 技术演进背景

随着大语言模型(如GPT-4、Llama-3-70B、DeepSeek-V3等)在企业级应用中的广泛部署，云原生(Kubernetes)平台已成为承载AI推理服务的事实标准。这一演进的核心驱动力在于：

1. **异构硬件资源池化**：通过Kubernetes Device Plugins统一管理H100、A100等高端GPU以及消费级显卡，实现算力的标准化交付。

2. **潮汐混部提升利用率**：利用优先级调度机制，白天运行在线推理服务，夜间运行离线训练任务，最大化降低GPU闲置成本。GPU小时成本高达数十元，资源利用率每提升1%即可产生显著经济效益。

3. **容器化解决环境依赖**：标准化PyTorch、CUDA等复杂依赖，简化跨集群部署。

### 2.2 当前MaaS平台架构与核心矛盾

**MaaS (Model-as-a-Service)平台现状**：

- **底层架构**：基于Kubernetes集群 + Volcano/Kueue批处理调度器 + RDMA/InfiniBand高性能网络
- **服务模式**：
  - PD分离架构（Prefill-Decode Disaggregation）：将推理任务拆分为计算密集型的Prefill实例和显存密集型的Decode实例，通过KV Cache传输协议通信
  - 多模型共存：单集群混合部署多个模型，或利用LoRA技术实现单底座多业务支持（Multi-LoRA Serving）

**本质矛盾**：

Kubernetes原生调度器(kube-scheduler)为**无状态(Stateless)应用**设计，而LLM推理本质上是**重状态(Stateful)**的服务。这种状态不仅包含：
- 模型权重本身(数十GB至数百GB)
- 请求动态生成的**KV Cache**(占用大量显存，决定推理效率)
- 微调的**Adapter权重**(LoRA场景)
- **请求间的路由上下文**(影响缓存命中率)

这种架构失配导致：在升级、扩缩容等运维操作时出现严重的**性能抖动**和**资源浪费**。

### 2.3 Fluid项目的技术优势

**Fluid**作为CNCF沙箱项目，是业界领先的Kubernetes原生分布式数据集编排和加速引擎，已在阿里巴巴、字节跳动等企业大规模应用。其核心能力包括：

- **数据集抽象与缓存加速**：通过统一的Dataset抽象，支持多级缓存(内存、SSD、远程存储)
- **弹性调度与亲和性**：支持数据亲和性调度，自动将计算调度到数据所在节点
- **可扩展Runtime插件**：已支持Alluxio、JuiceFS等多种分布式存储引擎

**技术契合点**：

Fluid在"数据-计算"协同调度方面的成熟能力，与LLM推理场景中"KV Cache-计算实例"的状态管理需求高度契合。本项目旨在将Fluid的数据编排能力扩展至**LLM推理状态管理**领域。

---

## 3. 解决的问题

本项目聚焦大模型推理服务在云原生平台上的**三大核心挑战**，这些挑战具有跨业务的共性，且技术方案具备学术创新性和产业引领性。

### 3.1 挑战一：升级场景下的零停机与性能保障

#### 问题描述

大模型服务具有极强的"状态"属性，与Kubernetes原生的滚动更新(Rolling Update)机制存在本质冲突：

1. **启动时延长**：加载数十GB的镜像和权重(如Llama-3-70B)在H100集群上通常耗时5-10分钟，包含：
   - 镜像拉取
   - 权重加载
   - CUDA Context初始化
   - torch.compile图捕获(Graph Capture)
   - KV Cache显存预分配

2. **性能回退(Cold Start)**：新Pod启动后，由于KV Cache未填充且计算图未编译，首批请求延迟极高(TTFT可达数秒)，导致用户体验骤降。

3. **拓扑依赖死锁**：在PD分离架构下，Prefill和Decode节点通过长连接(NCCL/gRPC)绑定。单独或无序升级某一角色，会导致协议不兼容或Pipeline断裂。

#### 研究方向

1. **原子化滚动更新(Atomic Rollout via RBG)**
   - 引入RoleBasedGroup (RBG) Operator，支持Gang Scheduling逻辑
   - 将配对的Prefill和Decode实例定义为不可分割的原子单元
   - 创建"影子组"(Shadow Group/Green Deployment)，只有当所有角色完成健康检查和预热后，才一次性切换流量

2. **模型热交换与显存检查点(CRIUgpu)**
   - 利用CRIUgpu技术实现GPU显存的检查点(Checkpoint)和恢复(Restore)
   - 构建"预热镜像"：将模型加载、编译、Cache预分配状态序列化，实现毫秒级冷启动
   - 支持模型多路复用(Multiplexing)：通过LRU缓存机制在显存中动态交换多个LoRA版本

3. **预测性预热(Predictive Warm-up)**
   - KV Cache预填充(Trace Replay)：利用历史热点请求记录，在新Pod启动后自动回放典型任务
   - 引入P99_Warmup_Latency指标：仅当新Pod的模拟预热延迟低于生产环境P99延迟的1.1倍时，才通过Readiness Probe
   - 基于Fluid的缓存预热能力，实现KV Cache的智能预加载

### 3.2 挑战二：扩容场景下的拓扑感知与实时决策

#### 问题描述

面对突发流量，如何快速获取资源并保证性能？传统CPU/GPU利用率指标在大模型场景下失效：

1. **可观测性Gap**：
   - CPU利用率无法反映GPU显存压力
   - GPU利用率无法区分Prefill(计算密集)与Decode(显存带宽密集)的不同瓶颈

2. **扩容策略困境**：
   - 是扩充Prefill算力以降低TTFT(Time to First Token)？
   - 还是扩充Decode算力以提升吞吐？
   - 如何平衡Prefill和Decode的协同性？

3. **物理拓扑约束**：
   - PD分离架构中，Prefill需向Decode传输大量KV Cache(GB级)
   - 跨ToR交换机通信会引入10倍以上延迟
   - Tensor Parallelism(TP)场景下，需将多卡调度到同一NVLink Domain

#### 研究方向

1. **Token动力学驱动的扩容决策**
   - 放弃CPU/GPU利用率，转向Token Velocity指标：
     - Prefill Velocity：输入处理速度(受FLOPS限制)
     - Decode Velocity：输出生成速度(受HBM带宽限制)
   - 监控KV Cache利用率(vllm:gpu_cache_usage_perc)作为核心信号
   - 结合Pending Queue Depth实现多级扩容触发机制

2. **分层扩容策略(Tier-based Scaling)**
   - Tier 1(秒级响应)：利用K8s PriorityClass机制，抢占低优先级训练任务
   - Tier 2(分钟级增长)：调用云API购买新节点，利用Image Streaming和P2P权重加载加速冷启动
   - 与Fluid的弹性伸缩能力集成，实现数据缓存与计算资源的协同扩展

3. **拓扑感知调度(Topology-Aware Placement)**
   - 扩容时调度器感知物理网络拓扑(Node -> S1 -> S2交换机层级)
   - 将配对的Prefill和Decode Pod调度到同一ToR交换机或NVLink Domain
   - 基于Fluid的亲和性调度能力，扩展支持GPU拓扑感知

### 3.3 挑战三：缩容场景下的状态保留与冷启动优化

#### 问题描述

对于长尾服务(Long-tail Services)，请求稀疏但持续时间长。直接删除Pod意味着：

1. **丢弃昂贵的KV Cache**：数GB的KV Cache需要数分钟重新计算
2. **算力浪费**：重复计算导致GPU利用率低下
3. **延迟劣化**：用户再次访问时需等待冷启动(Cold Start)

#### 研究方向

1. **KV Cache存算分离架构**
   - 构建GPU HBM -> CPU RAM -> Local SSD -> Remote Object Store (S3/Redis)的多级缓存体系
   - 缩容 = 释放计算节点(GPU)，保留状态(KV Cache写入远端存储)
   - 基于Fluid的分布式缓存能力，实现KV Cache的持久化和快速加载

2. **KV Cache迁移(Live Migration)**
   - 在缩容决策触发时，先通过RDMA网络将热点KV Cache迁移到活跃Pod
   - 支持基于状态感知的缩容策略，确保"人走茶不凉"

3. **Serverless LLM热启动**
   - 服务再次被唤醒时，新Pod直接从远端LMCache拉取KV Cache
   - 实现热启动(Warm Start)，而非从零计算
   - 利用Fluid的Dataset预加载机制，优化KV Cache的加载路径

### 3.4 跨领域技术创新点

本项目的创新性在于**将数据编排技术迁移至AI推理状态管理领域**，形成以下技术突破：

1. **状态即数据(State-as-Data)范式**
   - 将KV Cache、Adapter权重等运行时状态抽象为Fluid Dataset
   - 复用Fluid的多级缓存、亲和性调度、弹性伸缩等成熟能力
   - 实现状态的持久化、迁移、版本管理

2. **双重亲和性(Dual-Affinity)调度算法**
   - 同时考虑KV Cache命中率和LoRA权重加载状态
   - 在Multi-LoRA场景下优化请求路由决策
   - 扩展Fluid的数据亲和性调度至GPU拓扑感知

3. **联邦预调度层(Federated Pre-Scheduling)**
   - 在资源调度阶段捆绑Prefill-Decode Deployment Group
   - 基于拓扑资源树(Node -> S1 -> S2)维护全局最优布局
   - 与Fluid的调度框架深度集成

---

## 4. 期望合作方交付的指标

### 4.1 理论研究成果

#### 4.1.1 论文发表
- **目标**：在AI系统领域顶级会议/期刊发表高水平论文1-2篇
- **推荐会议**：OSDI、SOSP、ATC、EuroSys、MLSys、SoCC
- **推荐期刊**：ACM TOCS、IEEE TPDS、ACM TOS
- **核心贡献**：
  - 提出状态感知调度的形式化模型与理论证明
  - 双重亲和性调度算法的性能边界分析
  - 存算分离架构下的成本-性能权衡模型

#### 4.1.2 技术报告
- **开源技术白皮书**：详细描述系统架构、算法设计、实现细节(20-30页)
- **最佳实践指南**：面向工业界的部署指南(中英文)

### 4.2 系统实现与集成

#### 4.2.1 核心组件开发
以下组件需与Fluid项目深度集成，提交至fluid-cloudnative/fluid仓库：

1. **LLMRuntime插件** (核心交付物)
   - 实现Fluid Runtime接口，支持vLLM、SGLang、TensorRT-LLM等推理引擎
   - 提供KV Cache状态感知与管理能力
   - 支持LoRA Adapter的动态加载与缓存
   - **代码量估计**：5000-8000行Go代码
   - **交付标准**：通过Fluid社区Code Review，合并至主分支

2. **状态感知调度器扩展**
   - 扩展Kubernetes Scheduler Framework
   - 实现拓扑感知的Prefill-Decode协同调度
   - 支持双重亲和性(KV Cache + LoRA)的请求路由
   - **代码量估计**：3000-5000行Go代码
   - **交付标准**：作为Fluid Scheduler Plugin发布

3. **存算分离缓存层**
   - 基于Fluid的分布式缓存抽象实现KV Cache持久化
   - 支持GPU HBM -> CPU RAM -> SSD -> S3的多级缓存
   - 提供LMCache兼容的API接口
   - **代码量估计**：2000-3000行Go代码
   - **交付标准**：作为Fluid CacheEngine插件发布

#### 4.2.2 性能优化工具
- **KV Cache预热工具**：基于历史Trace自动生成预热脚本
- **拓扑可视化Dashboard**：展示GPU拓扑、KV Cache分布、请求路由路径
- **性能分析工具**：集成Prometheus Exporter，提供Token Velocity、Cache Hit Rate等指标

### 4.3 性能基准测试

#### 4.3.1 测试环境
- **硬件配置**：
  - GPU集群：≥16节点，每节点8×A100/H100 GPU
  - 网络：RDMA/InfiniBand(200Gbps+)
  - 存储：NVMe SSD + S3兼容对象存储
- **软件栈**：
  - Kubernetes 1.28+
  - Fluid v1.0.8+
  - vLLM/SGLang最新版本

#### 4.3.2 核心性能指标

| 指标类别 | 指标名称 | 基线(Baseline) | 目标改进 | 验证场景 |
|---------|---------|---------------|---------|---------|
| **升级场景** | Downtime | 30-60秒 | **≤1秒** | Llama-3-70B滚动更新 |
| | TTFT Jitter | ±200% | **≤20%** | 升级期间首字延迟波动 |
| | KV Cache Hit Rate Drop | 60-80% | **≤10%** | 升级前后缓存命中率跌幅 |
| **扩容场景** | Scale-out Latency(抢占) | N/A | **≤10秒** | 从决策到Pod Ready |
| | Scale-out Latency(新建) | 5-10分钟 | **≤2分钟** | 包含镜像拉取+权重加载 |
| | Topology Violation Rate | 30-50% | **≤5%** | PD配对被跨ToR调度的比例 |
| **缩容场景** | Cache Preservation Rate | 0% | **≥90%** | 缩容后KV Cache保留比例 |
| | Warm Start Speedup | 1× | **≥5×** | 相对于Cold Start的加速比 |
| | Storage Overhead | N/A | **≤20%** | 远端存储占用/GPU显存比例 |
| **综合性能** | P99 Latency | 基线 | **≤1.2×** | 在混合负载下保持性能 |
| | GPU Utilization | 40-60% | **≥75%** | 平均GPU利用率(含混部) |
| | Cost per 1M Tokens | 基线 | **-30%** | 综合TCO降低($/Token) |

#### 4.3.3 对比基线
- **Vanilla Kubernetes**：原生K8s调度 + 标准Rolling Update
- **Ray Serve**：业界主流LLM Serving框架
- **HeteroScale(字节)**：最新学术成果(ASPLOS'25)

### 4.4 开源社区贡献

#### 4.4.1 代码贡献
- **主仓库PR**：向fluid-cloudnative/fluid提交至少5个Major PR
- **代码质量**：
  - 单元测试覆盖率 ≥80%
  - 集成测试覆盖核心场景(升级/扩缩容)
  - 符合Fluid项目的Code Style和Architecture Guidelines

#### 4.4.2 文档与布道
- **官方文档**：撰写LLM推理场景的用户指南和开发者文档(中英文)
- **技术博客**：在Fluid官方博客或CNCF博客发布技术文章2-3篇
- **社区分享**：在Fluid社区会议、KubeCon/CloudNativeCon等会议进行技术分享

#### 4.4.3 案例验证
- **生产验证**：在至少1个真实业务场景中部署验证(如内部MaaS平台)
- **用户反馈**：收集至少3个企业用户的使用反馈和改进建议

### 4.5 知识产权

#### 4.5.1 专利申请
- **目标**：提交中国发明专利申请2-3项，涵盖：
  - 状态感知调度算法
  - KV Cache存算分离架构
  - 双重亲和性路由策略

#### 4.5.2 开源协议
- **所有代码遵循Apache 2.0协议**，贡献至fluid-cloudnative/fluid
- **论文预印本**：在arXiv公开发布，无版权限制

### 4.6 项目管理与交付节点

#### 4.6.1 项目周期
- **总周期**：12个月
- **阶段划分**：
  - **Phase 1 (M1-M3)**：需求分析、架构设计、原型开发
  - **Phase 2 (M4-M8)**：核心组件开发、性能优化、集成测试
  - **Phase 3 (M9-M12)**：生产验证、论文撰写、开源发布

#### 4.6.2 季度交付物
- **Q1**：技术方案设计文档 + LLMRuntime原型
- **Q2**：完整系统实现 + 初步性能测试报告
- **Q3**：优化版本 + 生产环境验证 + 论文初稿
- **Q4**：开源发布 + 最终性能报告 + 论文投稿

#### 4.6.3 验收标准
- **必选项(P0)**：
  - 核心性能指标达到目标(Downtime≤1s、TTFT Jitter≤20%)
  - 代码合并至Fluid主分支
  - 至少1篇会议论文投稿

- **加分项(P1)**：
  - 顶会论文接收
  - 多个企业用户采用
  - CNCF官方认可(如TOC推荐)

---

## 5. 项目价值与影响

### 5.1 技术引领性
本项目将**云原生数据编排**与**AI系统优化**两大前沿领域深度融合，填补Kubernetes在LLM推理场景下的能力空白，预期成果具备：
- **国际学术影响力**：目标发表OSDI/SOSP等顶会论文
- **产业标准潜力**：有望成为Kubernetes SIG-Scheduling的参考实现

### 5.2 商业价值
- **成本降低**：GPU利用率提升15-35%，每年节省数百万GPU成本
- **性能提升**：升级零停机、扩容延迟降低80%，直接改善用户体验
- **规模化支撑**：支持集团MaaS平台服务化能力，承载千亿级Token日吞吐

### 5.3 生态贡献
- **增强Fluid竞争力**：使Fluid成为业界首个原生支持LLM推理的数据编排系统
- **推动CNCF生态**：为Kubernetes AI Workload提供最佳实践
- **产学研融合**：建立高校-企业联合创新机制，培养AI系统方向人才

---

## 6. 研究团队期望

### 6.1 核心研究方向
- **分布式系统**：具有分布式调度、资源管理的深厚背景
- **AI系统优化**：熟悉LLM推理引擎(vLLM/TensorRT-LLM)和GPU编程
- **云原生技术**：深入理解Kubernetes架构和Scheduler Framework

### 6.2 合作模式
- **开放式协作**：定期召开技术评审会(双周/月度)
- **资源支持**：提供GPU集群测试环境、真实业务数据
- **导师制度**：指派集团AI Infra专家作为技术顾问

---

## 7. 总结

本项目聚焦**云原生大模型推理的状态管理与智能调度**这一极具挑战性和前瞻性的课题，旨在通过将Fluid的数据编排能力扩展至LLM推理领域，解决升级、扩缩容场景下的性能抖动和资源浪费问题。项目成果将：

1. **填补技术空白**：Kubernetes原生支持LLM Stateful Workload
2. **产出顶级论文**：目标OSDI/SOSP等旗舰会议
3. **开源回馈社区**：增强Fluid和CNCF生态
4. **创造商业价值**：降低GPU成本、提升用户体验

我们期待与全球顶尖研究者合作，共同推动AI基础设施技术的创新与突破！

---

**项目联系人**：Fluid社区技术委员会  
**联系方式**：fluid.opensource.project@gmail.com  
**项目官网**：https://github.com/fluid-cloudnative/fluid
