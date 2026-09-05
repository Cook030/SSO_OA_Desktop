import { Space, type SpaceProps } from "antd";
import type React from "react";

type SpaceOrientation = NonNullable<SpaceProps["orientation"]>;

export interface MhSpaceProps extends SpaceProps {
  /** @deprecated Use `orientation` instead. */
  direction?: SpaceOrientation;
  mhProps?: {
    __placeholder?: never;
  };
}

export const MhSpace: React.FC<MhSpaceProps> & {
  Compact: typeof Space.Compact;
} = ({ direction, orientation, ...restProps }) => {
  return <Space orientation={orientation ?? direction} {...restProps} />;
};

MhSpace.Compact = Space.Compact;

export default MhSpace;
