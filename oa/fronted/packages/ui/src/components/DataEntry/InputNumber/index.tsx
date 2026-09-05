import { InputNumber, type InputNumberProps } from "antd";
import type React from "react";

export interface MhInputNumberProps extends InputNumberProps {
  mhProps?: {
    __placeholder?: never;
  };
}

export const MhInputNumber: React.FC<MhInputNumberProps> = ({ ...restProps }) => {
  return <InputNumber {...restProps} />;
};

export default MhInputNumber;
