/*
 * Dataset Events component
 */

import React from 'react';
import { useCacheStore as useStore } from '@ks-console/shared';
import FluidEvents from '../../../../components/FluidEvents';

const DatasetEvents = () => {
  const [props] = useStore('DatasetDetailProps');
  const { detail, module } = props;

  return (
    <FluidEvents
      detail={detail}
      module={module || 'datasets'}
      resourceType="dataset"
      loadingText="Loading dataset details..."
    />
  );
};

export default DatasetEvents;