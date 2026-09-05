import { Spin, type SpinProps } from "antd";
import type React from "react";

export interface MhSpinProps extends SpinProps {
  mhProps?: {
    __placeholder?: never;
  };
}

export const MhSpin: React.FC<MhSpinProps> = ({ ...restProps }) => {
  return <Spin {...restProps} />;
};

export default MhSpin;
