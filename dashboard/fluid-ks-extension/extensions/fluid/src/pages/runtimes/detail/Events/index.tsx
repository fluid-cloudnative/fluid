/*
 * Runtime Events component
 */

import React from 'react';
import { useCacheStore as useStore } from '@ks-console/shared';
import FluidEvents from '../../../../components/FluidEvents';

const RuntimeEvents = () => {
  const [props] = useStore('RuntimeDetailProps');
  const { detail, module } = props;

  return (
    <FluidEvents
      detail={detail}
      module={module || 'runtimes'}
      resourceType="runtime"
      loadingText="Loading runtime details..."
    />
  );
};

export default RuntimeEvents;
