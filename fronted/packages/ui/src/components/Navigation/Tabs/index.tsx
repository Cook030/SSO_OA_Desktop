import { Tabs, type TabsProps } from "antd";
import type React from "react";

export interface MhTabsProps extends TabsProps {
  mhProps?: {
    __placeholder?: never;
  };
}

export const MhTabs: React.FC<MhTabsProps> & {
  TabPane: typeof Tabs.TabPane;
} = ({ ...restProps }) => {
  return <Tabs {...restProps} />;
};

MhTabs.TabPane = Tabs.TabPane;

export default MhTabs;
