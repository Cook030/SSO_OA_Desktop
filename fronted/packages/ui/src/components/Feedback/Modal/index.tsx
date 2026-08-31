import { Modal, type ModalProps } from "antd";
import type React from "react";

export interface MhModalProps extends ModalProps {
  mhProps?: {
    __placeholder?: never;
  };
}

export const MhModal: React.FC<MhModalProps> & {
  info: typeof Modal.info;
  success: typeof Modal.success;
  error: typeof Modal.error;
  warning: typeof Modal.warning;
  confirm: typeof Modal.confirm;
  destroyAll: typeof Modal.destroyAll;
  useModal: typeof Modal.useModal;
} = ({ ...restProps }) => {
  return <Modal {...restProps} />;
};

MhModal.info = Modal.info;
MhModal.success = Modal.success;
MhModal.error = Modal.error;
MhModal.warning = Modal.warning;
MhModal.confirm = Modal.confirm;
MhModal.destroyAll = Modal.destroyAll;
MhModal.useModal = Modal.useModal;

export default MhModal;
