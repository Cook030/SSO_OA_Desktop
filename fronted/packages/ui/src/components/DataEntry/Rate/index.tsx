import { Rate, type RateProps } from "antd";
import type React from "react";

export interface MhRateProps extends RateProps {
  mhProps?: {
    __placeholder?: never;
  };
}

export const MhRate: React.FC<MhRateProps> = ({ ...restProps }) => {
  return <Rate {...restProps} />;
};

export default MhRate;
