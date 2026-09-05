import { ColorPicker, type ColorPickerProps } from "antd";
import type React from "react";

export interface MhColorPickerProps extends ColorPickerProps {
  mhProps?: {
    __placeholder?: never;
  };
}

export const MhColorPicker: React.FC<MhColorPickerProps> = ({ ...restProps }) => {
  return <ColorPicker {...restProps} />;
};

export default MhColorPicker;
