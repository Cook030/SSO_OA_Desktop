import { FloatButton, type FloatButtonProps } from "antd";
import type React from "react";

export interface MhFloatButtonProps extends FloatButtonProps {
  // 预留扩展
  __placeholder?: never;
  // mhProps?: {};
}

export const MhFloatButton: React.FC<MhFloatButtonProps> & {
  Group: typeof FloatButton.Group;
  BackTop: typeof FloatButton.BackTop;
} = ({ ...restProps }) => {
  return <FloatButton {...restProps} />;
};

MhFloatButton.Group = FloatButton.Group;
MhFloatButton.BackTop = FloatButton.BackTop;

export default MhFloatButton;
