import { Anchor, type AnchorProps } from "antd";
import type React from "react";

export interface MhAnchorProps extends AnchorProps {
  mhProps?: {
    __placeholder?: never;
  };
}

export const MhAnchor: React.FC<MhAnchorProps> & {
  Link: typeof Anchor.Link;
} = ({ ...restProps }) => {
  return <Anchor {...restProps} />;
};

MhAnchor.Link = Anchor.Link;

export default MhAnchor;
