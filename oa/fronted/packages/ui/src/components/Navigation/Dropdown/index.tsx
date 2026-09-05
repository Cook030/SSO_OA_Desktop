import { Dropdown, type DropdownProps } from "antd";
import type React from "react";

export interface MhDropdownProps extends DropdownProps {
  mhProps?: {
    __placeholder?: never;
  };
}

export const MhDropdown: React.FC<MhDropdownProps> & {
  Button: typeof Dropdown.Button;
} = ({ ...restProps }) => {
  return <Dropdown {...restProps} />;
};

MhDropdown.Button = Dropdown.Button;

export default MhDropdown;
