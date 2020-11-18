# JVM Performance analysis tool usage

Sometimes, it's necessary to perform performance analysis on Fluid's underlying cache engine, such as Alluxio, to find performance bottlenecks more quickly.
[async-profiler](https://github.com/jvm-profiling-tools/async-profiler)is a comprehensive JVM profiling tool，supporting sampling of various events such as `cpu` and `lock`.
In this document, we introduce the simple usage of `async-profiler`.
Please query [async-profiler official document](https://github.com/jvm-profiling-tools/async-profiler) to get a more detailed tutorial.


## Download and unzip
   ```bash 
   $ wget https://github.com/jvm-profiling-tools/async-profiler/releases/download/v1.7.1/async-profiler-1.7.1-linux-x64.tar.gz
   $ tar -zxf async-profiler-1.7.1-linux-x64.tar.gz
   ``` 
    
## Simple usage

1. View Java process

    ```bash
    $ jps
    8576 AlluxioFuse
    33916 Jps
    ```
   
2. Sampling CPU time being taken up 
    ```bash
    $ ./profiler.sh -e cpu -i 1ms -d 300 -f cpu.txt <PID>
    ```

    **Description of command **:
    - `-e <EVENT>`: specify the sampling time, supporting `cpu`, `lock`, etc.
    - `-i <INTERVAL>`: specify the sampling interval, supporting the minimum nanosecond level sampling, but it is not recommended to set too small, because it will have a performance impact on the process being sampled
    - `-d <DURATION>`: specify the sampling duration, in seconds
    - `-f <FILENAME>`: specify the name and format of the saved file. The file format affects the presentation of the final result. The `txt` file format is convenient for viewing on the server, and the `svg` vector diagram format can be opened in the browser
    - `<PID>`: all samples must specify the PID of a JVM process, such as the PID of AlluxioFuse process in the `jps` command is `8576` 

3. Sampling result analysis

    async-profiler will periodically sample the JVM call stack and sort it from high to low according to the number of times it was sampled.
    It can be considered that, the higher the sampling times of function, the more CPU time it takes. Therefore, by observing call of the top function, you can quickly find the main performance bottleneck of the Java process.
        
    The following is the result of cpu sampling on AlluxioFuse. It consists of three parts：
        
    The first part introduces the number of sampling, GC conditions, etc. Among them, `Frame buffer usage` represents the buffer usage rate for saving the sampling results.
    If the interval of sampling is too small or the time of sampling is too long, it may prompt `overflow`. At this time, you can use `-b N` to increase the buffer capacity.
        
    The second part is the call stack of functions sorted from high to low according to the number of samples. The performance analysis focuses on the top few.
        
    The third part, at the end, is a function that CPU time being taken up from high to low.
    
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
> - General performance analysis only needs to sample `cpu` and `lock`, their results are more meaningful for reference
> - If it is memory-related tuning, try sampling the `alloc` event
> - The `wall` event samples the wall time. The `-t` option allows each process to be sampled separately. The combination of the two is better
> - Only one event can be sampled in the same process at the same time
