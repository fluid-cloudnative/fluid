JindoCache（前身为JindoFSx）是阿里云提供的云原生数据湖加速产品，支持数据缓存、元数据缓存等加速功能。JindoCache 能够为不同 CacheSet 提供不同的读写策略，满足数据湖的不同使用场景对访问加速的需求。
使用场景

JindoCache 可以用于如下场景：

- OLAP（Presto查询），提高查询性能，减少查询时间

- DataServing（HBase），显著降低P99延迟，减少request费用

- 大数据分析（Hive/Spark 报表），减少报表产出时间，优化计算集群成本

- 湖仓一体，减少request费用，优化catalog延迟

- AI，加速训练等场景，减少AI集群使用成本，提供更全面的能力支持

CacheSet

CacheSet 是 JindoCache 的缓存抽象。在实际使用中，往往不是所有数据都需要缓存加速。数据湖的计算种类繁多，场景丰富，有些数据可以使用激进的元数据缓存策略，有些数据完全不需要缓存。因此，JindoCache使用CacheSet作为配置粒度，给用户提供细粒度的访问策略选择。
缓存策略

- JindoCache 目前支持的数据读策略包括分布式数据缓存、一致性哈希数据缓存、本地缓存。

- JindoCache 目前支持可选的元数据缓存能力。