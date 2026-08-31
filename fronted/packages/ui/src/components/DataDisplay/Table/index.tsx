import { Table, type TableProps } from "antd";

export interface MhTableProps<T> extends TableProps<T> {
  mhProps?: {
    __placeholder?: never;
  };
}

export const MhTable = <T extends object>({ ...restProps }: MhTableProps<T>) => {
  return <Table {...restProps} />;
};

export default MhTable;
