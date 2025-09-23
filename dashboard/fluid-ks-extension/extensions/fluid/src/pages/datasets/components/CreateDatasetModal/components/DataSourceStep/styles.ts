import styled from 'styled-components';

export const StepContainer = styled.div`
  padding: 24px;
  min-height: 400px;
`;

export const MountItem = styled.div`
  border: 1px solid #e3e9ef;
  border-radius: 4px;
  padding: 16px;
  margin-bottom: 16px;
  background-color: #f9fbfd;
  position: relative;
`;

export const RemoveButton = styled.button`
  position: absolute;
  top: 8px;
  right: 8px;
  background: none;
  border: none;
  color: #ca2621;
  cursor: pointer;
  padding: 4px;
  border-radius: 4px;
  
  &:hover {
    background-color: #fff2f2;
  }
`;

export const AddMountButton = styled.button`
  background: none;
  border: 1px dashed #d8dee5;
  color: #3385ff;
  padding: 12px 16px;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  
  &:hover {
    border-color: #3385ff;
    background-color: #f8faff;
  }
`;

export const OptionsContainer = styled.div`
  margin-top: 16px;
`;

export const FormLabel = styled.label`
  display: block;
  margin-bottom: 8px;
  font-weight: 600;
`;

export const OptionalLabel = styled.span`
  color: #79879c;
  font-weight: 400;
  margin-left: 8px;
`;

export const SwitchContainer = styled.div`
  margin-bottom: 16px;
  display: flex;
  align-items: center;
  gap: 12px;
`;

export const SwitchLabel = styled.label`
  font-weight: 600;
  display: flex;
`;
