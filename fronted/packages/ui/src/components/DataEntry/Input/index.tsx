import { Input } from "antd";
import type React from "react";

type AntdInputProps = React.ComponentProps<typeof Input>;
type AntdTextAreaProps = React.ComponentProps<typeof Input.TextArea>;
type AntdSearchProps = React.ComponentProps<typeof Input.Search>;
type AntdPasswordProps = React.ComponentProps<typeof Input.Password>;
type InputElement = HTMLInputElement | HTMLTextAreaElement;

const { TextArea, Search, Password } = Input;

type MhInputBehaviorProps = {
  trimOnBlur?: boolean;
};

const createTrimmedBlurHandler = <T extends InputElement>(
  onBlur?: React.FocusEventHandler<T>,
  onChange?: React.ChangeEventHandler<T>,
  trimOnBlur = true
): React.FocusEventHandler<T> => {
  return event => {
    if (!trimOnBlur) {
      onBlur?.(event);
      return;
    }

    const trimmedValue = event.target.value.trim();

    if (event.target.value !== trimmedValue) {
      event.target.value = trimmedValue;
      event.currentTarget.value = trimmedValue;
      onChange?.(event as unknown as React.ChangeEvent<T>);
    }

    onBlur?.(event);
  };
};

const MhTextArea: React.FC<AntdTextAreaProps & MhInputBehaviorProps> = ({
  onBlur,
  onChange,
  trimOnBlur = true,
  ...restProps
}) => {
  return (
    <TextArea {...restProps} onChange={onChange} onBlur={createTrimmedBlurHandler(onBlur, onChange, trimOnBlur)} />
  );
};

const MhSearch: React.FC<AntdSearchProps & MhInputBehaviorProps> = ({
  onBlur,
  onChange,
  trimOnBlur = true,
  ...restProps
}) => {
  return <Search {...restProps} onChange={onChange} onBlur={createTrimmedBlurHandler(onBlur, onChange, trimOnBlur)} />;
};

const MhPassword: React.FC<AntdPasswordProps & MhInputBehaviorProps> = ({
  onBlur,
  onChange,
  trimOnBlur = true,
  ...restProps
}) => {
  return (
    <Password {...restProps} onChange={onChange} onBlur={createTrimmedBlurHandler(onBlur, onChange, trimOnBlur)} />
  );
};

export interface MhInputProps extends AntdInputProps {
  trimOnBlur?: boolean;
  mhProps?: {
    __placeholder?: never;
  };
}

export const MhInput: React.FC<MhInputProps> & {
  TextArea: React.FC<AntdTextAreaProps & MhInputBehaviorProps>;
  Search: React.FC<AntdSearchProps & MhInputBehaviorProps>;
  Password: React.FC<AntdPasswordProps & MhInputBehaviorProps>;
} = ({ onBlur, onChange, trimOnBlur = true, ...restProps }) => {
  return <Input {...restProps} onChange={onChange} onBlur={createTrimmedBlurHandler(onBlur, onChange, trimOnBlur)} />;
};

MhInput.TextArea = MhTextArea;
MhInput.Search = MhSearch;
MhInput.Password = MhPassword;

export default MhInput;
