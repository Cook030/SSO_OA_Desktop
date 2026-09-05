import { Checkbox, type CheckboxProps } from "antd";
import type React from "react";

export interface MhCheckboxProps extends CheckboxProps {
  mhProps?: {
    label?: string;
  };
}

export const MhCheckbox: React.FC<MhCheckboxProps> & {
  Group: typeof Checkbox.Group;
} = ({ mhProps, children, ...restProps }) => {
  return <Checkbox {...restProps}>{mhProps?.label || children}</Checkbox>;
};

MhCheckbox.Group = Checkbox.Group;

export default MhCheckbox;
