/*
 * DataLoad Events component
 */

import React from 'react';
import { useCacheStore as useStore } from '@ks-console/shared';
import FluidEvents from '../../../../components/FluidEvents';

const DataLoadEvents = () => {
  const [props] = useStore('DataLoadDetailProps');
  const { detail, module } = props;

  return (
    <FluidEvents
      detail={detail}
      module={module || 'dataloads'}
      resourceType="dataload"
      loadingText="Loading dataload details..."
    />
  );
};

export default DataLoadEvents;
