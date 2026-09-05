import { Alert, type AlertProps } from "antd";
import type React from "react";

export interface MhAlertProps extends AlertProps {
  mhProps?: {
    __placeholder?: never;
  };
}

export const MhAlert: React.FC<MhAlertProps> & {
  ErrorBoundary: typeof Alert.ErrorBoundary;
} = ({ ...restProps }) => {
  return <Alert {...restProps} />;
};

MhAlert.ErrorBoundary = Alert.ErrorBoundary;

export default MhAlert;
