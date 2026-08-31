import { Steps, type StepsProps } from "antd";
import type React from "react";

export interface MhStepsProps extends StepsProps {
  mhProps?: {
    __placeholder?: never;
  };
}

export const MhSteps: React.FC<MhStepsProps> = ({ ...restProps }) => {
  return <Steps {...restProps} />;
};

export default MhSteps;
