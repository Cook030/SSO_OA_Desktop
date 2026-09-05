import { Popconfirm } from "antd";
import type React from "react";

type AntdPopconfirmProps = React.ComponentProps<typeof Popconfirm>;

export interface MhPopconfirmProps extends AntdPopconfirmProps {
  mhProps?: {
    __placeholder?: never;
  };
}

export const MhPopconfirm: React.FC<MhPopconfirmProps> = ({ ...restProps }) => {
  return <Popconfirm {...restProps} />;
};

export default MhPopconfirm;
