import { Form, type FormProps } from "antd";
import type React from "react";

export interface MhFormProps extends FormProps {
  mhProps?: {
    __placeholder?: never;
  };
}

export const MhForm: React.FC<MhFormProps> & {
  Item: typeof Form.Item;
  List: typeof Form.List;
  ErrorList: typeof Form.ErrorList;
  Provider: typeof Form.Provider;
  useForm: typeof Form.useForm;
  useFormInstance: typeof Form.useFormInstance;
  useWatch: typeof Form.useWatch;
} = ({ children, ...restProps }) => {
  return <Form {...restProps}>{children as React.ReactNode}</Form>;
};

MhForm.Item = Form.Item;
MhForm.List = Form.List;
MhForm.ErrorList = Form.ErrorList;
MhForm.Provider = Form.Provider;
MhForm.useForm = Form.useForm;
MhForm.useFormInstance = Form.useFormInstance;
MhForm.useWatch = Form.useWatch;

export default MhForm;
