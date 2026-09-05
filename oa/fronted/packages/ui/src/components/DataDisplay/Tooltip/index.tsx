import type { TooltipProps } from "antd";
import { Tooltip } from "antd";
import type React from "react";

export interface MhTooltipProps extends TooltipProps {
  mhProps?: {
    __placeholder?: never;
  };
}

export const MhTooltip: React.FC<MhTooltipProps> = ({ ...restProps }) => {
  return <Tooltip {...restProps} />;
};

export default MhTooltip;
