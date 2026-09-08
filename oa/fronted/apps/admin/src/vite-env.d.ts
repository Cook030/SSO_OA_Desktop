/// <reference types="vite/client" />

declare global {
  interface ImportMetaEnv {
    /** WebSocket 端点，默认 /api/v1/realtime/ws */
    readonly VITE_WS_BASE_URL?: string;
  }
}

declare module "*.svg?react" {
  import type { FC, SVGProps } from "react";
  const ReactComponent: FC<SVGProps<SVGSVGElement>>;
  export default ReactComponent;
}
