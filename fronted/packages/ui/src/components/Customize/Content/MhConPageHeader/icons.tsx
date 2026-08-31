import type React from "react";

/** 页面头部首页图标 */
export const MhConPageHeaderHomeIcon: React.FC<React.SVGProps<SVGSVGElement>> = props => (
  <svg width="13" height="14" viewBox="0 0 13 14" fill="none" xmlns="http://www.w3.org/2000/svg" aria-hidden {...props}>
    <path
      d="M0 7.46485C0 6.79341 0.26529 6.14917 0.738093 5.67242L4.55627 1.82242C5.55202 0.818383 7.17525 0.818382 8.171 1.82242L11.9892 5.67242C12.462 6.14917 12.7273 6.79341 12.7273 7.46485V11.4545C12.7273 12.8604 11.5876 14 10.1818 14H2.54546C1.13964 14 0 12.8604 0 11.4545V7.46485Z"
      fill="#8A8A8A"
    />
    <rect x="5.09106" y="7.63672" width="2.54545" height="5.09091" rx="1.27273" fill="white" />
  </svg>
);

/** 收藏星标 */
export const MhConPageHeaderStarIcon: React.FC<
  React.SVGProps<SVGSVGElement> & { filled?: boolean; color?: string }
> = ({ filled = false, color, className, style, ...props }) => (
  <svg
    viewBox="0 0 24 24"
    fill={filled ? "currentColor" : "none"}
    stroke="currentColor"
    strokeWidth="1.5"
    className={className}
    style={color ? { color, ...style } : style}
    width="14"
    height="14"
    aria-hidden
    {...props}
  >
    <path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z" />
  </svg>
);

/** 面包屑分隔箭头 */
export const MhConBreadcrumbSeparatorIcon: React.FC<React.SVGProps<SVGSVGElement>> = props => (
  <svg width="12" height="12" fill="none" xmlns="http://www.w3.org/2000/svg" {...props}>
    <path
      d="M4.5 2L8.5 6L4.5 10"
      stroke="currentColor"
      strokeWidth="1.2"
      strokeLinecap="round"
      strokeLinejoin="round"
    />
  </svg>
);
