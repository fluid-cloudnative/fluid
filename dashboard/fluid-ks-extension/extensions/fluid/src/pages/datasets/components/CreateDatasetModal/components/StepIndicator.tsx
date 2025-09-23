import React from 'react';
import styled from 'styled-components';
import { Check } from '@kubed/icons';

declare const t: (key: string, options?: any) => string;

interface StepIndicatorProps {
  steps: Array<{
    key: string;
    title: string;
    description: string;
    optional?: boolean;
  }>;
  currentStep: number;
  completedSteps: Set<number>;
}

const StepContainer = styled.div`
  display: flex;
  align-items: stretch;
  padding: 16px 24px;
  background-color: #fff;
`;

const StepItem = styled.div<{ isActive: boolean; isCompleted: boolean }>`
  display: flex;
  align-items: center;
  flex: 1;
  position: relative;
  padding: 12px 16px;
  margin-right: 8px;
  border-radius: 8px 8px 0 0;
  cursor: pointer;
  transition: all 0.2s ease;
  background-color: ${props => {
    if (props.isActive) return '#fff';
    if (props.isCompleted) return '#f0f9ff';
    return 'transparent';
  }};
  border: ${props => {
    if (props.isActive) return '1px solid #e3e9ef';
    return '1px solid transparent';
  }};
  border-bottom: ${props => {
    if (props.isActive) return '1px solid #fff';
    return '1px solid #e3e9ef';
  }};

  &:last-child {
    margin-right: 0;
  }

  &:hover {
    background-color: ${props => props.isActive ? '#fff' : '#f0f9ff'};
  }
`;

const StepIcon = styled.div<{ isActive: boolean; isCompleted: boolean }>`
  width: 24px;
  height: 24px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: ${props => {
    if (props.isCompleted) return '#00aa72';
    if (props.isActive) return '#3385ff';
    return '#d8dee5';
  }};
  color: white;
  font-size: 12px;
  font-weight: 600;
  margin-right: 12px;
  flex-shrink: 0;
`;

const StepContent = styled.div`
  flex: 1;
  min-width: 0;
`;

const StepTitle = styled.div<{ isActive: boolean; isCompleted: boolean }>`
  font-size: 14px;
  font-weight: 600;
  color: ${props => {
    if (props.isActive) return '#3385ff';
    if (props.isCompleted) return '#00aa72';
    return '#79879c';
  }};
  margin-bottom: 2px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
`;

const StepDescription = styled.div<{ isActive: boolean; isCompleted: boolean }>`
  font-size: 12px;
  color: ${props => {
    if (props.isActive || props.isCompleted) return '#79879c';
    return '#c1c9d1';
  }};
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
`;

const OptionalBadge = styled.span`
  font-size: 10px;
  color: #79879c;
  background-color: #f0f9ff;
  padding: 2px 6px;
  border-radius: 10px;
  margin-left: 6px;
`;

const StepIndicator: React.FC<StepIndicatorProps> = ({
  steps,
  currentStep,
  completedSteps,
}) => {
  return (
    <StepContainer>
      {steps.map((step, index) => {
        const isActive = index === currentStep;
        const isCompleted = completedSteps.has(index);

        return (
          <StepItem
            key={step.key}
            isActive={isActive}
            isCompleted={isCompleted}
          >
            <StepIcon isActive={isActive} isCompleted={isCompleted}>
              {isCompleted ? (
                <Check size={14} />
              ) : (
                <span>{index + 1}</span>
              )}
            </StepIcon>
            <StepContent>
              <StepTitle isActive={isActive} isCompleted={isCompleted}>
                {t(step.title)}
                {step.optional && (
                  <OptionalBadge>{t('OPTIONAL')}</OptionalBadge>
                )}
              </StepTitle>
              <StepDescription isActive={isActive} isCompleted={isCompleted}>
                {t(step.description)}
              </StepDescription>
            </StepContent>
          </StepItem>
        );
      })}
    </StepContainer>
  );
};

export default StepIndicator;
