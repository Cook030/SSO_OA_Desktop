import { Masonry, type MasonryProps } from "antd";
import React from "react";

export interface MhMasonryProps<ItemDataType = any> extends MasonryProps<ItemDataType> {
  mhProps?: {
    __placeholder?: never;
  };
}

type MhMasonryRef = React.ElementRef<typeof Masonry>;

export const MhMasonry = React.forwardRef<MhMasonryRef, MhMasonryProps>(({ ...restProps }, ref) => {
  return <Masonry ref={ref} {...restProps} />;
});

MhMasonry.displayName = "MhMasonry";

export default MhMasonry;
