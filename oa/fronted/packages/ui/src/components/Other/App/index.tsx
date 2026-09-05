import { App, type AppProps } from "antd";
import type React from "react";

export interface MhAppProps extends AppProps {
  mhProps?: {
    __placeholder?: never;
  };
}

export const MhApp: React.FC<MhAppProps> & {
  useApp: typeof App.useApp;
} = ({ ...restProps }) => {
  return <App {...restProps} />;
};

MhApp.useApp = App.useApp;

export default MhApp;
