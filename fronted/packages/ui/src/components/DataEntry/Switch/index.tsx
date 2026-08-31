import { Switch } from "antd";
import type React from "react";

type AntdSwitchProps = React.ComponentProps<typeof Switch>;

export interface MhSwitchProps extends AntdSwitchProps {
  mhProps?: {
    __placeholder?: never;
  };
}

export const MhSwitch: React.FC<MhSwitchProps> = ({ ...restProps }) => {
  return <Switch {...restProps} />;
};

export default MhSwitch;
