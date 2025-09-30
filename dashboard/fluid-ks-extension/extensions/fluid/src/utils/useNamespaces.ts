import { useState, useEffect } from 'react';
import { request } from './request';

interface UseNamespacesResult {
  namespaces: string[];
  isLoading: boolean;
  error: string | null;
  refetchNamespaces: () => void;
}

export const useNamespaces = (currentCluster: string): UseNamespacesResult => {
  const [namespaces, setNamespaces] = useState<string[]>([]);
  const [isLoading, setIsLoading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);
  const [triggerRefetch, setTriggerRefetch] = useState<number>(0); // 用于手动触发刷新

  const fetchNamespaces = async () => {
    try {
      setIsLoading(true);
      setError(null);
      const response = await request('/api/v1/namespaces');

      if (!response.ok) {
        throw new Error(`${response.status}: ${response.statusText}`);
      }
      
      const data = await response.json();
      if (data && data.items) {
        const namespaceNames = data.items.map((item: any) => item.metadata.name);
        setNamespaces(namespaceNames);
      }
    } catch (err) {
      console.error('Failed to fetch namespaces:', err);
      setError(err instanceof Error ? err.message : String(err));
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchNamespaces();
  }, [currentCluster, triggerRefetch]); // 添加currentCluster和triggerRefetch依赖，集群切换或手动触发时重新获取命名空间

  const refetchNamespaces = () => {
    setTriggerRefetch(prev => prev + 1);
  };

  return { namespaces, isLoading, error, refetchNamespaces };
};
