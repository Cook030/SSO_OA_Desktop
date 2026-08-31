import { Watermark, type WatermarkProps } from "antd";
import type React from "react";

export interface MhWatermarkProps extends WatermarkProps {
  mhProps?: {
    __placeholder?: never;
  };
}

export const MhWatermark: React.FC<MhWatermarkProps> = ({ ...restProps }) => {
  return <Watermark {...restProps} />;
};

export default MhWatermark;
