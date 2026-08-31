import { Empty, type EmptyProps } from "antd";
import type React from "react";

export interface MhEmptyProps extends EmptyProps {
  mhProps?: {
    __placeholder?: never;
  };
}

export const MhEmpty: React.FC<MhEmptyProps> & {
  PRESENTED_IMAGE_DEFAULT: typeof Empty.PRESENTED_IMAGE_DEFAULT;
  PRESENTED_IMAGE_SIMPLE: typeof Empty.PRESENTED_IMAGE_SIMPLE;
} = ({ ...restProps }) => {
  return <Empty {...restProps} />;
};

MhEmpty.PRESENTED_IMAGE_DEFAULT = Empty.PRESENTED_IMAGE_DEFAULT;
MhEmpty.PRESENTED_IMAGE_SIMPLE = Empty.PRESENTED_IMAGE_SIMPLE;

export default MhEmpty;
