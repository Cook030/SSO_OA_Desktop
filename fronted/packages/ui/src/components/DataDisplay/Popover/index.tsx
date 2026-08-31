import { Popover, type PopoverProps } from "antd";
import type React from "react";

export interface MhPopoverProps extends PopoverProps {
  mhProps?: {
    __placeholder?: never;
  };
}

export const MhPopover: React.FC<MhPopoverProps> = ({ ...restProps }) => {
  return <Popover {...restProps} />;
};

export default MhPopover;
