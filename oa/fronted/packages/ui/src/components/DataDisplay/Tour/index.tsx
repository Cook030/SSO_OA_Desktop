import { Tour } from "antd";
import type React from "react";

type TourPropsBase = React.ComponentProps<typeof Tour>;

export interface MhTourProps extends TourPropsBase {
  mhProps?: {
    __placeholder?: never;
  };
}

export const MhTour: React.FC<MhTourProps> = ({ ...restProps }) => {
  return <Tour {...restProps} />;
};

export default MhTour;
