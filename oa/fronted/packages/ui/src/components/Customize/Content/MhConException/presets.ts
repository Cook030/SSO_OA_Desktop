export type MhConExceptionStatus = "403" | "404" | "500";

export interface MhConExceptionPreset {
  title: string;
  description: string;
}

export const MH_CON_EXCEPTION_PRESETS: Record<MhConExceptionStatus, MhConExceptionPreset> = {
  "403": {
    title: "403状态",
    description: "抱歉，您没有权限访问此页面"
  },
  "404": {
    title: "404状态",
    description: "抱歉，您访问的页面不存在"
  },
  "500": {
    title: "500状态",
    description: "哎呀！服务器好像出了点问题~"
  }
};
