// packages/ui/src/components/Message/index.tsx
import { message } from "antd";
import type React from "react";

// 定义 Props（无实际用途，仅为类型占位）
export interface MhMessageProps {
  mhProps?: {
    __placeholder?: never;
  };
}

// 创建一个空组件（不渲染任何内容）
export const MhMessage: React.FC<MhMessageProps> & {
  success: typeof message.success;
  error: typeof message.error;
  info: typeof message.info;
  warning: typeof message.warning;
  config: typeof message.config;
  destroy: typeof message.destroy;
} = () => {
  return null; // 不渲染 UI
};

// 挂载所有静态方法
MhMessage.success = message.success;
MhMessage.error = message.error;
MhMessage.info = message.info;
MhMessage.warning = message.warning;
MhMessage.config = message.config;
MhMessage.destroy = message.destroy;

export default MhMessage;
