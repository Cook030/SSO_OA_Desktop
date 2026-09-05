import { Carousel, type CarouselProps } from "antd";
import type React from "react";

export interface MhCarouselProps extends CarouselProps {
  mhProps?: {
    __placeholder?: never;
  };
}

export const MhCarousel: React.FC<MhCarouselProps> = ({ ...restProps }) => {
  return <Carousel {...restProps} />;
};

export default MhCarousel;
