import { Divider, type DividerProps } from "antd";
import type React from "react";

export interface MhDividerProps extends DividerProps {
  mhProps?: {
    __placeholder?: never;
  };
}

export const MhDivider: React.FC<MhDividerProps> = ({ ...restProps }) => {
  return <Divider {...restProps} />;
};

export default MhDivider;
