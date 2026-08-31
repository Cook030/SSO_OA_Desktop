import type { MhConExceptionStatus } from "@mh-repo/ui";
import type React from "react";
import type { MhIconName } from "../../../components/SvgComp";
import SvgComp from "../../../components/SvgComp";
import styles from "./ExceptionIllustration.module.less";

const statusIconMap: Record<MhConExceptionStatus, MhIconName> = {
  "403": "mh-403",
  "404": "mh-404",
  "500": "mh-500"
};

export interface ExceptionIllustrationProps {
  status: MhConExceptionStatus;
}

export const ExceptionIllustration: React.FC<ExceptionIllustrationProps> = ({ status }) => (
  <SvgComp name={statusIconMap[status]} className={styles.illustration} aria-hidden />
);
