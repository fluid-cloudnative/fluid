# Demo - Alluxio Tieredstore Configuration

[Alluxio](https://github.com/Alluxio/alluxio) is one of the distributed cache engines leveraged by Fluid.
It supports tieredstores to store cached data in different location, for example different directories with different storage types.
By appropriate configurations on tieredstores, users can get better I/O throughput from Fluid and eliminate bottlenecks when accessing data.

The guide introduces you how to configure Alluxio's tieredstore in a declarative way in Fluid.

To get more tech detail about Alluxio's tieredstore, please refer to [Cache-related docs offered by Alluxio](https://docs.alluxio.io/ee/user/stable/en/core-services/Caching.html?q=tieredstore#configuring-alluxio-storage)

## Prerequisites

- [Fluid](https://github.com/fluid-cloudnative/fluid)(version >= 0.3.0)

Please refer to [Fluid installation documentation](https://github.com/fluid-cloudnative/fluid/blob/master/docs/en/userguide/install.md) to complete installation.

## Single-Tier Configuration

Here is an typical example for an AlluxioRuntime:

```yaml
apiVersion: data.fluid.io/v1alpha1
kind: AlluxioRuntime
metadata:
  name: dataset
spec:
  ...
  tieredstore:
    levels:
      - path: /dev/shm
        mediumtype: MEM
        quota: 2Gi
```

`spec.tieredstore.levels` contains only one level, so Alluxio will run with single tieredstore.

A brief description for each property involved in the above-mentioned example is as follows:
- `path`: where data cache actually stores
- `mediumtype`: one of the three values(`MEM`, `SSD`, `HDD`), used to specify medium for `path`
- `quota`: maximium cache capacity for the level

## Single-Tier Multi-Directory Configuration

The best way to demonstrate such configuration is to give an example like this:

```yaml
apiVersion: data.fluid.io/v1alpha1
kind: AlluxioRuntime
metadata:
  name: dataset
spec:
  ...
  tieredstore:
    levels:
      - path: /mnt/ssd0/cache,/mnt/ssd1/cache
        mediumtype: SSD
        quota: 4Gi
```

To use multiple directories as  Alluxio's tieredstore, 
the only difference is to add more directories in `path` with comma(",") as their separator.
Take the yaml above as an example, with `path` containing both "/mnt/ssd0/cache" and "/mnt/ssd1/cache", 
Alluxio will use these two directories as its cache store in the meantime. 

The example also implies some best practices about when you might want to use such configuration: If there is a bottleneck
introduced by storage device itself(e.g. limited by Hard disk I/O throughput), using multiple storage devices("/mnt/ssd0" and "/mnt/ssd1" in the example above) as Alluxio's tieredstore
can reduce load on each device and get a higher I/O throughput.

> Note: For now, Fluid only support tieredstores with homogeneous medium type. 
> That is, it is not allowed to use different hybrid storages in a tieredstore in Fluid.

Also please note that, if multi-directory tieredstore is enabled, `quota` will be divided equally to each directory. 
Take the yaml above as an example, with `quota` setting to "4Gi", each directory in `path` will have a cache capacity of "2Gi"(i.e. 4Gi / 2)

If that's not your desired way, Fluid also provides `quotaList` to configure cache capacity for each directory:

```yaml
apiVersion: data.fluid.io/v1alpha1
kind: AlluxioRuntime
metadata:
  name: dataset
spec:
  properties:
    ...
    # [default property in fluid]
    # alluxio.worker.allocator.class: alluxio.worker.block.allocator.MaxFreeAllocator
  ...
  tieredstore:
    levels:
      - path: /mnt/ssd0/cache,/mnt/ssd1/cache
        mediumtype: SSD
        #quota: 4Gi
        quotaList: 3Gi,2Gi
```

`quotaList` allows you to set maximum cache capacity for each directory. 
The `quotaList` will distributed its values into directories in the same order you defined in `path`, 
so the number of `quotaList` must be in consistent with the number of `path`. For example, with the yaml above, 
"/mnt/ssd0/cache" has a maximum cache capacity of "3Gi", while "/mnt/ssd1/cache" has capacity of "2Gi".

Another Alluxio property related to multi-directory tieredstore is `alluxio.worker.allocator.class`.
All the supported values are described as follows:
- "MaxFreeAllocator": Always try to allocate the cache to the storage directory that currently has the most availability
- "RoundRobinAllocator": On each tier, maintain a Round Robin order of storage directories. Try to allocate the new cache into a directory following the Round Robin order, and if that does not work, go to the next lower tier.
- "GreedyAllocator": Always try to put the new cache into the first directory that can contain it.

By default, Fluid uses "MaxFreeAllocator" to decide where to store a new cache. Users can feel free to change this behavior by setting corresponding allocator to the Alluxio property.
For example, users can change `alluxio.worker.allocator.class` to "alluxio.worker.block.allocator.RoundRobinAllocator" to choose round robin strategy.

## Multi-Tier Configuration

Here is an example for multi-tier configuration:

```yaml
apiVersion: data.fluid.io/v1alpha1
kind: AlluxioRuntime
metadata:
  name: dataset
spec:
  ...
  tieredstore:
    levels:
      - path: /dev/shm
        mediumtype: MEM
        quota: 2Gi
      - path: /mnt/ssd0/cache,/mnt/ssd1/cache
        mediumtype: SSD
        quotaList: 3Gi,2Gi
```

Multiple tieredstores are just tieredstores with some order. Take the yaml above as an example, 
we specify two tieredstores: the first level uses memory for high-speed data access and the second one uses
SSD to get bigger cache capacity.

Level order defined in `spec.tieredstore.levels` will not affect the actual level order used by Alluxio.
Before Alluxio launched, Fluid will firstly sort the levels according to `mediumtype`, and storages with higher I/O throughput will get higher priority.
That is, Fluid will sort tieredstores in the following orders: "MEM" < "SSD" < "HDD".

> Note: Alluxio uses a different way to report its capacity usage when using multi-tier configuration. 
> For now, this will lead to some inaccuracy when showing `Dataset.Cached` and `Dataset.CachedPercentage` in Fluid.


