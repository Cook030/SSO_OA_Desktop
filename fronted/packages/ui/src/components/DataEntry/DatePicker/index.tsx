import { DatePicker } from "antd";
import type React from "react";

type AntdDatePickerProps = React.ComponentProps<typeof DatePicker>;

export interface MhDatePickerProps extends AntdDatePickerProps {
  mhProps?: {
    __placeholder?: never;
  };
}

type DatePickerComponent = React.FC<AntdDatePickerProps> & {
  RangePicker: typeof DatePicker.RangePicker;
};

export const MhDatePicker: DatePickerComponent = ({ ...restProps }) => {
  return <DatePicker {...restProps} />;
};

MhDatePicker.RangePicker = DatePicker.RangePicker;

export default MhDatePicker;
