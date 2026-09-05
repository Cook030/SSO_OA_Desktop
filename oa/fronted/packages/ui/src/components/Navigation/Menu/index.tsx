import { Menu, type MenuProps } from "antd";
import type React from "react";

export interface MhMenuProps extends MenuProps {
  mhProps?: {
    __placeholder?: never;
  };
}

export const MhMenu: React.FC<MhMenuProps> & {
  Item: typeof Menu.Item;
  SubMenu: typeof Menu.SubMenu;
  ItemGroup: typeof Menu.ItemGroup;
  Divider: typeof Menu.Divider;
} = ({ ...restProps }) => {
  return <Menu {...restProps} />;
};

MhMenu.Item = Menu.Item;
MhMenu.SubMenu = Menu.SubMenu;
MhMenu.ItemGroup = Menu.ItemGroup;
MhMenu.Divider = Menu.Divider;

export default MhMenu;
