import { Select, type SelectProps } from "antd";
import type React from "react";

export interface MhSelectProps extends SelectProps {
  mhProps?: {
    __placeholder?: never;
  };
}

export const MhSelect: React.FC<MhSelectProps> & {
  Option: typeof Select.Option;
  OptGroup: typeof Select.OptGroup;
} = ({ ...restProps }) => {
  return <Select {...restProps} />;
};

MhSelect.Option = Select.Option;
MhSelect.OptGroup = Select.OptGroup;

export default MhSelect;
