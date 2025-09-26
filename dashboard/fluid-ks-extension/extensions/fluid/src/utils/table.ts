/**
 * 转换请求参数
 * @param params 请求参数
 * @returns 转换后的请求参数
 */
export const transformRequestParams = (params: Record<string, any>) => {
    const { parameters = {}, pageIndex, filters = [], pageSize, sortBy } = params;
    console.log('转换前的请求参数:', params);
    
    // 从filters中获取搜索关键词
    const keyword = filters[0]?.value;
    
    // 构建查询参数
    const result: Record<string, any> = {
      ...parameters,
      limit: pageSize,
      page: pageIndex + 1,
    };
    
    // 如果有搜索关键词，添加name参数
    if (keyword) {
      result.name = keyword;
      console.log('添加搜索关键词:', keyword);
    }
  
    if (sortBy && sortBy.length > 0) {
      // sortBy=name&ascending=true
      console.log('sortBy', sortBy);
      const { id, desc } = sortBy[0];
      result.sortBy = id;
      result.ascending = !desc;
    }
    
    
    console.log('转换后的请求参数:', result);
    return result;
};


