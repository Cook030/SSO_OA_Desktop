import { QRCode, type QRCodeProps } from "antd";
import type React from "react";

export interface MhQRCodeProps extends QRCodeProps {
  mhProps?: {
    __placeholder?: never;
  };
}

export const MhQRCode: React.FC<MhQRCodeProps> = ({ ...restProps }) => {
  return <QRCode {...restProps} />;
};

export default MhQRCode;
