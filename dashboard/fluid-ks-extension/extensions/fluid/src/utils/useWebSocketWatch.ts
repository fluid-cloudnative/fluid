import { useEffect, useState, useRef } from 'react';
import { getWebSocketUrl } from './request';
import { DebouncedFunc } from 'lodash';

interface UseWebSocketWatchOptions {
  namespace: string;
  resourcePlural: string; // 例如：'dataloads', 'datasets', 'alluxioruntimes'
  currentCluster: string;
  debouncedRefresh: DebouncedFunc<() => void>; // 用于刷新数据表格的防抖函数
  onResourceDeleted?: () => void; // 可选回调，用于处理资源删除事件（例如清空选中项）
  refreshInterval?: number; // 轮询刷新间隔，默认为15000ms
  initialEventsWindow?: number; // 连接后跳过初始ADDED事件的时间窗口，默认为2000ms
}

export const useWebSocketWatch = ({
  namespace,
  resourcePlural,
  currentCluster,
  debouncedRefresh,
  onResourceDeleted,
  refreshInterval = 15000,
  initialEventsWindow = 2000,
}: UseWebSocketWatchOptions) => {
  const [wsConnected, setWsConnected] = useState<boolean>(false);
  const isComponentUnmounting = useRef(false); // 使用 ref 跟踪组件卸载状态

  useEffect(() => {
    // 根据命名空间和资源复数形式构建 WebSocket 路径
    const wsPath = namespace
      ? `/apis/data.fluid.io/v1alpha1/watch/namespaces/${namespace}/${resourcePlural}?watch=true`
      : `/apis/data.fluid.io/v1alpha1/watch/${resourcePlural}?watch=true`;
    const wsUrl = getWebSocketUrl(wsPath);

    console.log(`=== 启动 ${resourcePlural} WebSocket监听 ===`);
    console.log("WebSocket URL:", wsUrl);

    let ws: WebSocket | null = null;
    let reconnectTimeout: NodeJS.Timeout | undefined;
    let pollingInterval: NodeJS.Timeout | undefined;
    let reconnectCount = 0;
    const maxReconnectAttempts = 5;
    let connectionStartTime = 0;

    const connect = () => {

      try {
        console.log("连接WebSocket:", wsUrl);
        ws = new WebSocket(wsUrl);

        ws.onopen = () => {
          console.log(`=== ${resourcePlural} WebSocket连接成功 ===`);
          setWsConnected(true);
          reconnectCount = 0; // 重置重连计数
          connectionStartTime = Date.now(); // 记录连接建立时间

          // WebSocket连接成功，停止任何正在进行的轮询
          if (pollingInterval) {
            console.log("WebSocket连接成功，停止轮询");
            clearInterval(pollingInterval);
            pollingInterval = undefined;
          }
        };

        ws.onmessage = (event) => {
          console.log(`=== ${resourcePlural} WebSocket收到消息 ===`);
          try {
            const data = JSON.parse(event.data);
            console.log("消息类型:", data.type);
            console.log("对象名称:", data.object?.metadata?.name);

            // 处理 ADDED, DELETED, MODIFIED 事件
            if (['ADDED', 'DELETED', 'MODIFIED'].includes(data.type)) {
              // 跳过连接建立初期的 ADDED 事件，避免不必要的刷新
              if (data.type === 'ADDED' && connectionStartTime > 0) {
                const timeSinceConnection = Date.now() - connectionStartTime;
                if (timeSinceConnection < initialEventsWindow) {
                  console.log(`=== 跳过连接初期的ADDED事件 (连接后${timeSinceConnection}ms)，避免不必要的刷新 ===`);
                  return;
                }
              }

              console.log("=== 检测到数据变化，准备防抖刷新 ===");
              // 如果是 DELETED 事件，触发可选的删除回调（例如清空选中项）
              if (data.type === 'DELETED' && onResourceDeleted) {
                console.log("检测到删除事件，执行onResourceDeleted回调");
                onResourceDeleted();
              }
              // 使用防抖函数刷新表格
              debouncedRefresh();
            }
          } catch (e) {
            console.error("解析WebSocket消息失败:", e);
          }
        };

        ws.onclose = (event) => {
          console.log(`=== ${resourcePlural} WebSocket连接关闭 ===`, event.code, event.reason || '无reason', ws?.url);
          setWsConnected(false);

          // 如果组件未卸载且重连次数未达到上限，则尝试重连
          if (!isComponentUnmounting.current && reconnectCount < maxReconnectAttempts) {
            const delay = Math.min(1000 * Math.pow(2, reconnectCount), 10000); // 指数退避
            console.log(`${delay}ms后尝试重连 (${reconnectCount + 1}/${maxReconnectAttempts})`);

            reconnectTimeout = setTimeout(() => {
              reconnectCount++;
              connect();
            }, delay);
          } else if (!isComponentUnmounting.current && reconnectCount >= maxReconnectAttempts) {
            // 重连次数用完，启动轮询保底方案
            console.log("=== WebSocket重连失败，启动轮询保底方案 ===");
            if (!pollingInterval) { // 防止重复设置轮询
              pollingInterval = setInterval(() => {
                console.log("=== 执行轮询刷新 ===");
                debouncedRefresh();
              }, refreshInterval);
            }
          } else if (isComponentUnmounting.current) {
            console.log(`=== ${resourcePlural}组件卸载，正常关闭WebSocket ===`);
          }
        };

        ws.onerror = (error) => {
          console.error(`=== ${resourcePlural} WebSocket错误 ===`, error);
          setWsConnected(false);
          // 如果 WebSocket 发生错误，立即尝试 fallback 到轮询
          if (!pollingInterval && !isComponentUnmounting.current) {
            console.log("=== WebSocket错误，启动轮询保底方案 ===");
            pollingInterval = setInterval(() => {
              console.log("=== 执行轮询刷新 ===");
              debouncedRefresh();
            }, refreshInterval);
          }
        };
      } catch (error) {
        console.error(`=== 创建 ${resourcePlural} WebSocket失败 ===`, error);
        setWsConnected(false);
        // 如果 WebSocket 创建失败，立即 fallback 到轮询
        if (!pollingInterval && !isComponentUnmounting.current) {
          console.log("=== WebSocket创建失败，启动轮询保底方案 ===");
          pollingInterval = setInterval(() => {
            console.log("=== 执行轮询刷新 ===");
            debouncedRefresh();
          }, refreshInterval);
        }
      }
    };

    // 启动连接
    connect();

    // 清理函数，在组件卸载时执行
    return () => {
      console.log(`=== 清理 ${resourcePlural} WebSocket连接和轮询 ===`, ws?.url);
      isComponentUnmounting.current = true; // 设置卸载标志
      if (reconnectTimeout) {
        clearTimeout(reconnectTimeout);
      }
      if (pollingInterval) {
        clearInterval(pollingInterval);
      }
      if (ws) {
        ws.close(1000, 'Component unmounting');
      }
      // 取消防抖函数的待执行任务
      debouncedRefresh.cancel();
      setWsConnected(false);
    };
  }, [namespace, currentCluster, resourcePlural]);

  return { wsConnected };
};
