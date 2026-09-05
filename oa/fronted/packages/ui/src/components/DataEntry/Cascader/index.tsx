import { Cascader } from "antd";
import type React from "react";

type CascaderProps = React.ComponentProps<typeof Cascader>;

export type MhCascaderProps = CascaderProps & {
  mhProps?: {
    __placeholder?: never;
  };
};

export const MhCascader: React.FC<MhCascaderProps> = props => {
  const { mhProps, ...rest } = props;

  return <Cascader {...rest} />;
};

export default MhCascader;
