import { Flex, type FlexProps } from "antd";
import type React from "react";

export interface MhFlexProps extends FlexProps {
  mhProps?: {
    __placeholder?: never;
  };
}

export const MhFlex: React.FC<MhFlexProps> = ({ ...restProps }) => {
  return <Flex {...restProps} />;
};

export default MhFlex;
