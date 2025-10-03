/*
 * Shared styled components for ResourceStatus components
 * Used by both Runtime and Dataset ResourceStatus components
 */

import styled from 'styled-components';

// Card wrapper with consistent margin
export const CardWrapper = styled.div`
  margin-bottom: 12px;
`;

// Info grid for basic information display
export const InfoGrid = styled.div`
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 20px;
  padding: 20px;
`;

// Individual info item container
export const InfoItem = styled.div`
  display: flex;
  flex-direction: column;
  gap: 8px;
`;

// Label for info items
export const InfoLabel = styled.div`
  font-size: 12px;
  color: #79879c;
  font-weight: 600;
`;

// Value for info items
export const InfoValue = styled.div`
  font-size: 14px;
  color: #242e42;
  font-weight: 600;
`;

// Status card container
export const StatusCard = styled.div`
  background: #ffffff;
  border: 1px solid #e3e9ef;
  border-radius: 4px;
  padding: 20px;
  margin-bottom: 12px;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.04);
`;

// Status card header
export const StatusHeader = styled.div`
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 20px;
  padding-bottom: 16px;
  border-bottom: 1px solid #f0f0f0;
`;

// Icon container for status cards
export const StatusIcon = styled.div`
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  border-radius: 8px;
  background: #f0f9ff;
  color: #0369a1;
`;

// Title for status cards
export const StatusTitle = styled.div`
  font-size: 16px;
  font-weight: 600;
  color: #242e42;
`;

// Grid layout for status items
export const StatusGrid = styled.div`
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(120px, 1fr));
  gap: 24px;
`;

// Individual status item container
export const StatusItem = styled.div`
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 4px;
`;

// Value display for status items
export const StatusValue = styled.div`
  font-size: 20px;
  font-weight: 600;
  color: #242e42;
  line-height: 1.2;
`;

// Label for status items
export const StatusLabel = styled.div`
  font-size: 12px;
  color: #79879c;
  text-transform: uppercase;
  font-weight: 500;
  letter-spacing: 0.5px;
`;
