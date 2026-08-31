import { Badge, type BadgeProps } from "antd";
import type React from "react";

export interface MhBadgeProps extends BadgeProps {
  mhProps?: {
    __placeholder?: never;
  };
}

export const MhBadge: React.FC<MhBadgeProps> & {
  Ribbon: typeof Badge.Ribbon;
} = ({ ...restProps }) => {
  return <Badge {...restProps} />;
};

MhBadge.Ribbon = Badge.Ribbon;

export default MhBadge;
