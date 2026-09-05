import type React from "react";

export interface SvgCompProps {
  /** Ant Design 图标名或 mh 图标名（`mh-` 前缀） */
  name: string;
  className?: string;
  style?: React.CSSProperties;
  size?: number | string;
  spin?: boolean;
  filled?: boolean;
  active?: boolean;
  color?: string;
}

export type SvgIconComponent = React.ComponentType<Record<string, unknown>>;

export interface CreateSvgCompOptions {
  antdIconRegistry: Record<string, SvgIconComponent>;
  mhIconRegistry: Record<string, SvgIconComponent>;
  /** 开发环境未知图标时是否 console.warn，默认 true */
  warnOnUnknown?: boolean;
}

const mergeSizeStyle = (
  size: number | string | undefined,
  style?: React.CSSProperties
): React.CSSProperties | undefined => {
  if (size === undefined) return style;
  const dimension = typeof size === "number" ? `${size}px` : size;
  return { width: dimension, height: dimension, fontSize: dimension, ...style };
};

export const createSvgComp = ({ antdIconRegistry, mhIconRegistry, warnOnUnknown = true }: CreateSvgCompOptions) => {
  const isAntdIconName = (name: string): name is keyof typeof antdIconRegistry => name in antdIconRegistry;
  const isMhIconName = (name: string): name is keyof typeof mhIconRegistry => name in mhIconRegistry;

  const SvgComp: React.FC<SvgCompProps> = ({ name, className, style, size, spin, filled, active, color }) => {
    const mergedStyle = mergeSizeStyle(size, style);

    if (isMhIconName(name)) {
      const MhIcon = mhIconRegistry[name];
      if (name === "mh-star") {
        return <MhIcon className={className} style={mergedStyle} filled={filled} color={color} />;
      }
      if (name === "mh-info") {
        return <MhIcon className={className} style={mergedStyle} active={active} />;
      }
      return <MhIcon className={className} style={mergedStyle} aria-hidden />;
    }

    if (isAntdIconName(name)) {
      const AntdIcon = antdIconRegistry[name];
      return <AntdIcon className={className} style={mergedStyle} spin={spin} />;
    }

    if (warnOnUnknown) {
      console.warn(`[SvgComp] Unknown icon name: ${name}`);
    }
    return null;
  };

  const renderSvgIcon = (name: string, props?: Omit<SvgCompProps, "name">): React.ReactNode => (
    <SvgComp name={name} {...props} />
  );

  return {
    SvgComp,
    renderSvgIcon,
    isAntdIconName,
    isMhIconName,
    antdIconRegistry,
    mhIconRegistry
  };
};
