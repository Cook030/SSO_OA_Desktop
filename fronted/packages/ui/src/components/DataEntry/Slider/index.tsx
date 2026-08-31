import { Slider } from "antd";
import type React from "react";

// export interface MhSliderProps extends SliderSingleProps {
//   mhProps?: {
//     __placeholder?: never;
//   };
// }

export type MhSliderProps = React.ComponentProps<typeof Slider> & {
  mhProps?: {
    __placeholder?: never;
  };
};

export const MhSlider: React.FC<MhSliderProps> = ({ ...restProps }) => {
  return <Slider {...restProps} />;
};

export default MhSlider;
