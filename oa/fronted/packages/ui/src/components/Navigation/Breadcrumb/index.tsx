import { Breadcrumb, type BreadcrumbProps } from "antd";
import type React from "react";

export interface MhBreadcrumbProps extends BreadcrumbProps {
  mhProps?: {
    __placeholder?: never;
  };
}

export const MhBreadcrumb: React.FC<MhBreadcrumbProps> & {
  Item: typeof Breadcrumb.Item;
  Separator: typeof Breadcrumb.Separator;
} = ({ ...restProps }) => {
  return <Breadcrumb {...restProps} />;
};

MhBreadcrumb.Item = Breadcrumb.Item;
MhBreadcrumb.Separator = Breadcrumb.Separator;

export default MhBreadcrumb;
