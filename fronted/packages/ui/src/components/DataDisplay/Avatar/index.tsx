import { Avatar, type AvatarProps } from "antd";
import type React from "react";

export interface MhAvatarProps extends AvatarProps {
  __placeholder?: never;
}

export const MhAvatar: React.FC<MhAvatarProps> & {
  Group: typeof Avatar.Group;
} = ({ ...restProps }) => {
  return <Avatar {...restProps} />;
};

MhAvatar.Group = Avatar.Group;

export default MhAvatar;
