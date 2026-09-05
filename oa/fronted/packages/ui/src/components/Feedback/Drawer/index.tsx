import { Drawer, type DrawerProps } from "antd";
import type React from "react";

export interface MhDrawerProps extends DrawerProps {
  mhProps?: {
    __placeholder?: never;
  };
}

export const MhDrawer: React.FC<MhDrawerProps> = ({ ...restProps }) => {
  return <Drawer {...restProps} />;
};

export default MhDrawer;
