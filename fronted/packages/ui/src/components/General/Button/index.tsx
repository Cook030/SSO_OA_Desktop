import { Button, type ButtonProps } from "antd";
import type React from "react";

export interface MhButtonProps extends ButtonProps {
  mhProps?: {
    prefix?: string;
    suffix?: string;
  };
}

export const MhButton: React.FC<MhButtonProps> = ({ children, mhProps, ...restProps }) => {
  const content = (
    <>
      {mhProps?.prefix && <span style={{ marginRight: 4 }}>{mhProps?.prefix}</span>}
      {children}
      {mhProps?.suffix && <span style={{ marginLeft: 4 }}>{mhProps?.suffix}</span>}
    </>
  );

  return <Button {...restProps}>{content}</Button>;
};

export default MhButton;
