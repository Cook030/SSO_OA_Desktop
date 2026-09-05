import { Upload, type UploadProps } from "antd";
import type React from "react";

export interface MhUploadProps extends UploadProps {
  mhProps?: {
    __placeholder?: never;
  };
}

export const MhUpload: React.FC<MhUploadProps> & {
  Dragger: typeof Upload.Dragger;
} = ({ children, ...restProps }) => {
  return <Upload {...restProps}>{children}</Upload>;
};

MhUpload.Dragger = Upload.Dragger;

export default MhUpload;
