import { Affix, type AffixProps } from "antd";
import type React from "react";

export interface MhAffixProps extends AffixProps {
  mhProps?: {
    __placeholder?: never;
  };
}

export const MhAffix: React.FC<MhAffixProps> = ({ ...restProps }) => {
  return <Affix {...restProps} />;
};

export default MhAffix;
