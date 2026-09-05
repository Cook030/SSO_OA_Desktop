import { Radio, type RadioProps } from "antd";
import type React from "react";

export interface MhRadioProps extends RadioProps {
  mhProps?: {
    label?: string;
  };
}

export const MhRadio: React.FC<MhRadioProps> & {
  Group: typeof Radio.Group;
  Button: typeof Radio.Button;
} = ({ mhProps, children, className = "", ...restProps }) => {
  return (
    <Radio className={`custom-radio ${className}`} {...restProps}>
      {mhProps?.label || children}
    </Radio>
  );
};

MhRadio.Group = Radio.Group;
MhRadio.Button = Radio.Button;

export default MhRadio;
