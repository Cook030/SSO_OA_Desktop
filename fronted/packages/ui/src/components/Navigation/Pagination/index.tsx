import { Pagination, type PaginationProps } from "antd";
import type React from "react";

export interface MhPaginationProps extends PaginationProps {
  mhProps?: {
    __placeholder?: never;
  };
}

export const MhPagination: React.FC<MhPaginationProps> = ({ ...restProps }) => {
  return <Pagination {...restProps} />;
};

export default MhPagination;
