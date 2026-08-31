import { Progress, type ProgressProps } from "antd";
import type React from "react";

export interface MhProgressProps extends ProgressProps {
  mhProps?: {
    __placeholder?: never;
  };
}

export const MhProgress: React.FC<MhProgressProps> = ({ ...restProps }) => {
  return <Progress {...restProps} />;
};

export default MhProgress;
