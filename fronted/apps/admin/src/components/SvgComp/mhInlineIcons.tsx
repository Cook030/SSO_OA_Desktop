import type React from "react";

/** mh-filter：筛选漏斗 */
export const MhFilterIcon: React.FC<React.SVGProps<SVGSVGElement>> = props => (
  <svg viewBox="0 0 14 14" width="14" height="14" fill="none" aria-hidden="true" {...props}>
    <path
      d="M1.75 2.333h10.5L8.167 7v3.5l-2.334 1.167V7L1.75 2.333Z"
      fill="currentColor"
      opacity="0.75"
      stroke="currentColor"
      strokeLinejoin="round"
    />
  </svg>
);

/** mh-info：信息提示圆 */
export const MhInfoIcon: React.FC<React.SVGProps<SVGSVGElement> & { active?: boolean }> = ({
  active = false,
  ...props
}) => (
  <svg viewBox="0 0 16 16" width="16" height="16" fill="none" aria-hidden="true" {...props}>
    <circle cx="8" cy="8" r="6.25" stroke={active ? "#1677FF" : "#BFBFBF"} strokeWidth="1.2" />
    <path d="M8 7v3.2" stroke={active ? "#1677FF" : "#8C8C8C"} strokeWidth="1.2" strokeLinecap="round" />
    <circle cx="8" cy="4.7" r="0.8" fill={active ? "#1677FF" : "#8C8C8C"} />
  </svg>
);

/** mh-star：收藏星标 */
export const MhStarIcon: React.FC<React.SVGProps<SVGSVGElement> & { filled?: boolean; color?: string }> = ({
  filled = false,
  color,
  className,
  style,
  ...props
}) => (
  <svg
    viewBox="0 0 24 24"
    fill={filled ? "currentColor" : "none"}
    stroke="currentColor"
    strokeWidth="1.5"
    className={className}
    style={color ? { color, ...style } : style}
    width="14"
    height="14"
    aria-hidden="true"
    {...props}
  >
    <path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z" />
  </svg>
);

/** mh-copy：复制 */
export const MhCopyIcon: React.FC<React.SVGProps<SVGSVGElement>> = props => (
  <svg viewBox="0 0 16 16" width="16" height="16" fill="none" aria-hidden="true" {...props}>
    <path
      d="M6.667 4.667h6a1 1 0 0 1 1 1v7a1 1 0 0 1-1 1h-6a1 1 0 0 1-1-1v-7a1 1 0 0 1 1-1Z"
      stroke="currentColor"
      strokeWidth="1.2"
      strokeLinejoin="round"
    />
    <path
      d="M3.333 11.333h-.666a1 1 0 0 1-1-1v-7a1 1 0 0 1 1-1h6a1 1 0 0 1 1 1V4"
      stroke="currentColor"
      strokeWidth="1.2"
      strokeLinejoin="round"
    />
  </svg>
);

/** mh-breadcrumb-separator：面包屑分隔箭头 */
export const MhBreadcrumbSeparatorIcon: React.FC<React.SVGProps<SVGSVGElement>> = props => (
  <svg width="12" height="12" fill="none" xmlns="http://www.w3.org/2000/svg" className="text-black/25" {...props}>
    <path
      d="M4.5 2L8.5 6L4.5 10"
      stroke="currentColor"
      strokeWidth="1.2"
      strokeLinecap="round"
      strokeLinejoin="round"
    />
  </svg>
);
