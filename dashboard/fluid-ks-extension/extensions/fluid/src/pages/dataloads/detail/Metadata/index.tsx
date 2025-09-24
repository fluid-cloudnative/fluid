/*
 * DataLoad Metadata component
 */

import React from 'react';
import FluidMetadata from '../../../../components/FluidMetadata';

const Metadata = () => {
  return (
    <FluidMetadata
      storeKey="DataLoadDetailProps"
      loadingText="Loading dataload details..."
    />
  );
};

export default Metadata;
