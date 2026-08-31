import { Transfer, type TransferProps } from "antd";

export interface MhTransferProps<T> extends TransferProps<T> {
  __placeholder?: never;
}

export const MhTransfer = <T extends object>({ ...restProps }: MhTransferProps<T>) => {
  return <Transfer {...restProps} />;
};

export default MhTransfer;
