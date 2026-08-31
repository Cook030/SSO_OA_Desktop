import type React from "react";
import { MhModal, type MhModalProps } from "../../../Feedback/Modal";
import { MhButton } from "../../../General/Button";
import { CloseOutlined } from "../../../General/Icon";
import styles from "./index.module.less";

export interface MhConModalProps extends Omit<MhModalProps, "title" | "footer"> {
  title?: React.ReactNode;
  headerExtra?: React.ReactNode;
  footer?: React.ReactNode;
}

export interface MhConModalCloseButtonProps {
  onClick: () => void;
  className?: string;
}

export const MhConModalCloseButton: React.FC<MhConModalCloseButtonProps> = ({ onClick, className }) => (
  <MhButton
    type="text"
    className={[styles.closeButton, className].filter(Boolean).join(" ")}
    onClick={onClick}
    onMouseDown={event => event.preventDefault()}
    aria-label="关闭"
  >
    <CloseOutlined />
  </MhButton>
);

export const MhConModal: React.FC<MhConModalProps> = ({
  title,
  headerExtra,
  footer,
  children,
  className,
  centered = true,
  closable = false,
  ...restProps
}) => {
  const showHeader = title != null || headerExtra != null;
  const showFooter = footer != null;

  return (
    <MhModal
      {...restProps}
      centered={centered}
      closable={closable}
      footer={null}
      className={[styles.modal, className].filter(Boolean).join(" ")}
    >
      <div className={styles.wrapper}>
        {showHeader ? (
          <div className={[styles.header, headerExtra ? styles.headerWithExtra : ""].filter(Boolean).join(" ")}>
            <div className={styles.headerInner}>
              {title != null ? <div className={styles.title}>{title}</div> : null}
            </div>
          </div>
        ) : null}

        <div className={[styles.content, !showHeader ? styles.contentNoHeader : ""].filter(Boolean).join(" ")}>
          {children}
        </div>

        {showFooter ? <div className={styles.footer}>{footer}</div> : null}

        {headerExtra ? <div className={styles.headerExtraAnchor}>{headerExtra}</div> : null}
      </div>
    </MhModal>
  );
};

export default MhConModal;
