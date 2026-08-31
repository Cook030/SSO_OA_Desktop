import { Tag, type TagProps } from "antd";
import type React from "react";

export interface MhTagProps extends TagProps {
  mhProps?: {
    __placeholder?: never;
  };
}

export const MhTag: React.FC<MhTagProps> & {
  CheckableTag: typeof Tag.CheckableTag;
} = ({ ...restProps }) => {
  return <Tag {...restProps} />;
};

MhTag.CheckableTag = Tag.CheckableTag;

export default MhTag;
