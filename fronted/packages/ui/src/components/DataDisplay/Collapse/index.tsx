import { Collapse, type CollapseProps } from "antd";
import type React from "react";

export interface MhCollapseProps extends CollapseProps {
  mhProps?: {
    __placeholder?: never;
  };
}

export const MhCollapse: React.FC<MhCollapseProps> & {
  Panel: typeof Collapse.Panel;
} = ({ ...restProps }) => {
  return <Collapse {...restProps} />;
};

MhCollapse.Panel = Collapse.Panel;

export default MhCollapse;
