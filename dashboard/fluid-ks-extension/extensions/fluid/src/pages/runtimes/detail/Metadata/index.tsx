/*
 * Runtime Metadata component
 */

import React from 'react';
import FluidMetadata from '../../../../components/FluidMetadata';

const Metadata = () => {
  return (
    <FluidMetadata
      storeKey="RuntimeDetailProps"
      loadingText="Loading runtime details..."
    />
  );
};

export default Metadata;
