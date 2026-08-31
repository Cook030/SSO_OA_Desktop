import { TimePicker } from "antd";
import type React from "react";

type AntdTimePickerProps = React.ComponentProps<typeof TimePicker>;

export interface MhTimePickerProps extends AntdTimePickerProps {
  mhProps?: {
    __placeholder?: never;
  };
}

type TimePickerComponent = React.FC<AntdTimePickerProps> & {
  RangePicker: typeof TimePicker.RangePicker;
};

export const MhTimePicker: TimePickerComponent = ({ ...restProps }) => {
  return <TimePicker {...restProps} />;
};

MhTimePicker.RangePicker = TimePicker.RangePicker;

export default MhTimePicker;
