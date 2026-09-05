import type React from "react";
import illustration403 from "./assets/403.svg";
import illustration404 from "./assets/404.svg";
import illustration500 from "./assets/500.svg";
import type { MhConExceptionStatus } from "./presets";

const illustrationMap: Record<MhConExceptionStatus, string> = {
  "403": illustration403,
  "404": illustration404,
  "500": illustration500
};

export interface MhConExceptionIllustrationProps extends React.ImgHTMLAttributes<HTMLImageElement> {
  status: MhConExceptionStatus;
}

export const MhConExceptionIllustration: React.FC<MhConExceptionIllustrationProps> = ({
  status,
  className,
  alt = "",
  ...props
}) => (
  <img
    src={illustrationMap[status]}
    className={className}
    alt={alt}
    aria-hidden={alt === "" ? true : undefined}
    {...props}
  />
);
