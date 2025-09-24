/*
 * Dataset Metadata component
 */

import React from 'react';
import FluidMetadata from '../../../../components/FluidMetadata';

const Metadata = () => {
  return (
    <FluidMetadata
      storeKey="DatasetDetailProps"
      loadingText="Loading dataset details..."
    />
  );
};

export default Metadata;