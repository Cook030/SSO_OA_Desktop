import { Tree, type TreeDataNode, type TreeProps } from "antd";
import type React from "react";

export interface MhTreeProps extends TreeProps {
  mhProps?: {
    __placeholder?: never;
  };
}

export interface MhTreeDataNode extends TreeDataNode {
  mhProps?: {
    __placeholder?: never;
  };
}

export const MhTree: React.FC<MhTreeProps> & {
  TreeNode: typeof Tree.TreeNode;
  DirectoryTree: typeof Tree.DirectoryTree;
} = ({ ...restProps }) => {
  return <Tree {...restProps} />;
};

MhTree.TreeNode = Tree.TreeNode;
MhTree.DirectoryTree = Tree.DirectoryTree;

export default MhTree;
