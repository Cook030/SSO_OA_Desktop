import { Image, type ImageProps } from "antd";
import type React from "react";

export interface MhImageProps extends ImageProps {
  mhProps?: {
    __placeholder?: never;
  };
}

export const MhImage: React.FC<MhImageProps> & {
  PreviewGroup: typeof Image.PreviewGroup;
} = ({ ...restProps }) => {
  return <Image {...restProps} />;
};

MhImage.PreviewGroup = Image.PreviewGroup;

export default MhImage;
