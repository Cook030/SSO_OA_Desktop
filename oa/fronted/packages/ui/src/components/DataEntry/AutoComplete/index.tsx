import { AutoComplete, type AutoCompleteProps } from "antd";
import type React from "react";

export interface MhAutoCompleteProps extends AutoCompleteProps {
  mhProps?: {
    __placeholder?: never;
  };
}

export const MhAutoComplete: React.FC<MhAutoCompleteProps> = ({ ...restProps }) => {
  return <AutoComplete {...restProps} />;
};

export default MhAutoComplete;
