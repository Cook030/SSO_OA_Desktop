import { Segmented, type SegmentedProps } from "antd";
import type React from "react";

export interface MhSegmentedProps extends SegmentedProps {
  mhProps?: {
    __placeholder?: never;
  };
}

export const MhSegmented: React.FC<MhSegmentedProps> = ({ ...restProps }) => {
  return <Segmented {...restProps} />;
};

export default MhSegmented;
