import { TreeSelect, type TreeSelectProps } from "antd";
import type React from "react";

export interface MhTreeSelectProps extends TreeSelectProps {
  mhProps?: {
    __placeholder?: never;
  };
}

export const MhTreeSelect: React.FC<MhTreeSelectProps> & {
  TreeNode: typeof TreeSelect.TreeNode;
} = ({ ...restProps }) => {
  return <TreeSelect {...restProps} />;
};

MhTreeSelect.TreeNode = TreeSelect.TreeNode;

export default MhTreeSelect;
