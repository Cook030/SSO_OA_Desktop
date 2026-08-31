import { Result, type ResultProps } from "antd";
import type React from "react";

export interface MhResultProps extends ResultProps {
  mhProps?: {
    __placeholder?: never;
  };
}

export const MhResult: React.FC<MhResultProps> = ({ ...restProps }) => {
  return <Result {...restProps} />;
};

export default MhResult;
