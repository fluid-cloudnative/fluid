# JVM性能分析工具使用

有时需要对Fluid的底层缓存引擎，如Alluxio，进行性能分析，以更加迅速地找到性能瓶颈。
[async-profiler](https://github.com/jvm-profiling-tools/async-profiler)是一款非常全面的JVM profiling工具，支持对`cpu`，`lock`等多种事件采样。
在本文档中，我们介绍`async-profiler`的简单使用方法。
请参考[async-profiler官方文档](https://github.com/jvm-profiling-tools/async-profiler)获取更加详细的使用教程。

## 下载并解压
   ```bash
   $ wget https://github.com/jvm-profiling-tools/async-profiler/releases/download/v1.7.1/async-profiler-1.7.1-linux-x64.tar.gz
   $ tar -zxf async-profiler-1.7.1-linux-x64.tar.gz
   ```

## 简单使用

1. 查看Java进程
    ```bash
    $ jps
    8576 AlluxioFuse
    33916 Jps
    ```

2. 采样占用CPU时间
    ```bash
    $ ./profiler.sh -e cpu -i 1ms -d 300 -f cpu.txt <PID>
    ```

    **命令说明**:
    - `-e <EVENT>`： 指定要采样的时间，支持`cpu`, `lock`等
    - `-i <INTERVAL>`： 指定采样间隔， 支持最小纳秒级别采样，但不建议设置太小，对会被采样的进程造成较大的性能影响
    - `-d <DURATION>`: 指定采样持续时间，以秒为单位
    - `-f <FILENAME>`: 指定保存文件名字和格式，文件格式影响最终结果呈现形式，`txt`文档格式方便在服务器查看，`svg`矢量图格式可在浏览器打开
    - `<PID>`: 所有采样都必须指定一个JVM进程的PID， 如`jps`命令中的AlluxioFuse进程`8576`

3. 采样结果分析

    async-profiler会周期性地采样JVM调用栈，并按照被采样次数由高到低排序。
    可认为采样次数越高的函数，花费的CPU时间也是越多的。因此，通过观察排名前几位的函数调用情况，可快速找出Java进程的主要性能瓶颈。
    
    下面是对AlluxioFuse进行cpu采样的结果，由三部分内容组成。
    
    第一部分，介绍采样的次数，GC情况等，其中的`Frame buffer usage`表示保存采样结果的buffer使用率。
    如果采样间隔太小，或者时间太长，可能会提示`overflow`，此时可用`-b N`调大buffer容量。
    
    第二部分，是按照采样次数由高到低排序后的函数调用栈，性能分析时重点关注前几位。
    
    第三部分，在最后，是占用CPU时间由高到低的函数。
    ```bash
    $ cat cpu.txt
    --- Execution profile ---
    Total samples       : 131152
    GC_active           : 8 (0.01%)
    unknown_Java        : 321 (0.24%)
    not_walkable_Java   : 234 (0.18%)
    deoptimization      : 7 (0.01%)
    skipped             : 17 (0.01%)
    
    Frame buffer usage  : 23.9887%
    
    --- 9902352571 ns (7.48%), 9820 samples
      [ 0] copy_user_enhanced_fast_string_[k]
      [ 1] copyout_[k]
      [ 2] copy_page_to_iter_[k]
      [ 3] skb_copy_datagram_iter_[k]
      [ 4] tcp_recvmsg_[k]
      [ 5] inet_recvmsg_[k]
      [ 6] sock_read_iter_[k]
      [ 7] __vfs_read_[k]
      [ 8] vfs_read_[k]
      [ 9] ksys_read_[k]
      [10] do_syscall_64_[k]
      [11] entry_SYSCALL_64_after_hwframe_[k]
      [12] read
      [13] io.netty.channel.unix.FileDescriptor.readAddress
      [14] io.netty.channel.unix.FileDescriptor.readAddress
      [15] io.netty.channel.epoll.AbstractEpollChannel.doReadBytes
      [16] io.netty.channel.epoll.AbstractEpollStreamChannel$EpollStreamUnsafe.epollInReady
      [17] io.netty.channel.epoll.EpollEventLoop.processReady
      [18] io.netty.channel.epoll.EpollEventLoop.run
      [19] io.netty.util.concurrent.SingleThreadEventExecutor$4.run
      [20] io.netty.util.internal.ThreadExecutorMap$2.run
      [21] java.lang.Thread.run
    
    --- 6862595502 ns (5.19%), 6836 samples
      [ 0] jbyte_disjoint_arraycopy
      [ 1] io.netty.util.internal.PlatformDependent0.copyMemoryWithSafePointPolling
      [ 2] io.netty.util.internal.PlatformDependent0.copyMemory
      [ 3] io.netty.util.internal.PlatformDependent.copyMemory
      [ 4] io.netty.buffer.UnsafeByteBufUtil.getBytes
      [ 5] io.netty.buffer.PooledUnsafeDirectByteBuf.getBytes
      [ 6] io.netty.buffer.AbstractUnpooledSlicedByteBuf.getBytes
      [ 7] io.netty.buffer.AbstractByteBuf.readBytes
      [ 8] io.grpc.netty.NettyReadableBuffer.readBytes
      [ 9] io.grpc.internal.CompositeReadableBuffer$3.readInternal
      [10] io.grpc.internal.CompositeReadableBuffer$ReadOperation.read
      [11] io.grpc.internal.CompositeReadableBuffer.execute
      [12] io.grpc.internal.CompositeReadableBuffer.readBytes
      [13] alluxio.grpc.ReadableDataBuffer.readBytes
      [14] alluxio.client.block.stream.BlockInStream.readInternal
      [15] alluxio.client.block.stream.BlockInStream.read
      [16] alluxio.client.file.AlluxioFileInStream.read
      [17] alluxio.client.file.cache.LocalCacheFileInStream.readExternalPage
      [18] alluxio.client.file.cache.LocalCacheFileInStream.read
      [19] alluxio.fuse.AlluxioJniFuseFileSystem.read
      [20] alluxio.jnifuse.AbstractFuseFileSystem.readCallback
    
    --- 6257779616 ns (4.73%), 6234 samples
      [ 0] jlong_disjoint_arraycopy
      [ 1] java.util.HashMap.isEmpty
      [ 2] java.util.HashSet.isEmpty
      [ 3] alluxio.client.file.cache.store.MemoryPageStore.put
      [ 4] alluxio.client.file.cache.LocalCacheManager.addPage
      [ 5] alluxio.client.file.cache.LocalCacheManager.putInternal
      [ 6] alluxio.client.file.cache.LocalCacheManager.lambda$put$0
      [ 7] alluxio.client.file.cache.LocalCacheManager$$Lambda$324.504173880.run
      [ 8] java.util.concurrent.Executors$RunnableAdapter.call
      [ 9] java.util.concurrent.FutureTask.run
      [10] java.util.concurrent.ThreadPoolExecutor.runWorker
      [11] java.util.concurrent.ThreadPoolExecutor$Worker.run
      [12] java.lang.Thread.run
    
    --- 5784713511 ns (4.37%), 5730 samples
      [ 0] memcpy_erms_[k]
      [ 1] fuse_copy_do?[fuse]_[k]
      [ 2] fuse_copy_page?[fuse]_[k]
      [ 3] fuse_copy_args?[fuse]_[k]
      [ 4] fuse_dev_do_write?[fuse]_[k]
      [ 5] fuse_dev_write?[fuse]_[k]
      [ 6] do_iter_readv_writev_[k]
      [ 7] do_iter_write_[k]
      [ 8] vfs_writev_[k]
      [ 9] do_writev_[k]
      [10] do_syscall_64_[k]
      [11] entry_SYSCALL_64_after_hwframe_[k]
      [12] writev
    # ...
             ns  percent  samples  top
      ----------  -------  -------  ---
     18748768943   14.17%    18618  jlong_disjoint_arraycopy
     12465418383    9.42%    12375  copy_user_enhanced_fast_string_[k]
     11200001765    8.46%    11153  /usr/lib/jvm/java-8-openjdk-amd64/jre/lib/amd64/server/libjvm.so
      8891711765    6.72%     8863  jbyte_disjoint_arraycopy
      5934957181    4.48%     5877  memcpy_erms_[k]
      5887736474    4.45%     5766  _raw_spin_unlock_irqrestore_[k]
      5722725002    4.32%     5669  /lib/x86_64-linux-gnu/libc-2.23.so
      5029972109    3.80%     5024  jint_disjoint_arraycopy
      3313144887    2.50%     3284  _raw_spin_lock_[k]
      2345730045    1.77%     2341  jshort_disjoint_arraycopy
      2111075419    1.60%     2098  java.lang.Long.equals
      1844566888    1.39%     1838  __do_page_fault_[k]
      1493059573    1.13%     1472  get_user_pages_fast_[k]
      1371331539    1.04%     1340  finish_task_switch_[k]
      1333100221    1.01%     1324  clear_page_erms_[k]
      1165108284    0.88%     1155  io.netty.util.Recycler$WeakOrderQueue.transfer
      1010331386    0.76%     1004  __free_pages_ok_[k]
       806598842    0.61%      798  fuse_copy_do?[fuse]_[k]
       805970066    0.61%      802  try_charge_[k]
       773960647    0.58%      757  read
       660352796    0.50%      656  get_page_from_freelist_[k]
    ```
> **Tips**:
> - 一般性能分析只需要采样`cpu`和`lock`即可，它们的结果是比较有参考意义的
> - 如果是和内存相关的调优，可试着采样`alloc`事件
> - `wall`事件采样墙上时间，`-t`选项让每个进程分开采样，它俩搭配使用效果比较好
> - 同一进程同时只能采样一种事件
