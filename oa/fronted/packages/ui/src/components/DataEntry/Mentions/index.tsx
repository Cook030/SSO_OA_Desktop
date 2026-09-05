import { Mentions, type MentionsProps } from "antd";
import type React from "react";

export interface MhMentionsProps extends MentionsProps {
  mhProps?: {
    __placeholder?: never;
  };
}

export const MhMentions: React.FC<MhMentionsProps> & {
  Option: typeof Mentions.Option;
} = ({ ...restProps }) => {
  return <Mentions {...restProps} />;
};

MhMentions.Option = Mentions.Option;

export default MhMentions;
