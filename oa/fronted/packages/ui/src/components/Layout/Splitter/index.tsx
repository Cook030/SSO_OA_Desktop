import { Splitter } from "antd";
import type React from "react";

export type MhSplitterProps = React.ComponentProps<typeof Splitter> & {
  mhProps?: {
    __placeholder?: never;
  };
};

export const MhSplitter: React.FC<MhSplitterProps> & {
  Panel: typeof Splitter.Panel;
} = ({ ...restProps }) => {
  return <Splitter {...restProps} />;
};

MhSplitter.Panel = Splitter.Panel;

export default MhSplitter;
