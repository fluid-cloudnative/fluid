import React from 'react';
import { Navigate } from 'react-router-dom';
import App from '../App';
import DatasetList from '../pages/datasets/list';
import RuntimeList from '../pages/runtimes/list';
import DataLoadList from '../pages/dataloads/list';
import datasetDetailRoutes from '../pages/datasets/detail/routes';
import runtimeDetailRoutes from '../pages/runtimes/detail/routes';
import dataloadDetailRoutes from '../pages/dataloads/detail/routes';

export default [
  // 重定向根路径到默认集群
  {
    path: '/fluid',
    element: <Navigate to="/fluid/host/datasets" replace />,
  },
  {
    path: '/fluid/:cluster',
    element: <App />,
    children: [
      { index: true, element: <Navigate to="datasets" replace /> },
      {
        path: 'datasets',
        element: <DatasetList />,
      },
      {
        path: 'runtimes',
        element: <RuntimeList />,
      },
      {
        path: 'dataloads',
        element: <DataLoadList />,
      },
    ]
  },
  ...datasetDetailRoutes,
  ...runtimeDetailRoutes,
  ...dataloadDetailRoutes,
];
