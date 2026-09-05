import type { CalendarProps } from "antd";
import { Calendar } from "antd";
import type { Dayjs } from "dayjs";
import type React from "react";

export interface MhCalendarProps extends CalendarProps<Dayjs> {
  mhProps?: {
    __placeholder?: never;
  };
}

export const MhCalendar: React.FC<MhCalendarProps> = ({ ...restProps }) => {
  return <Calendar {...restProps} />;
};

export default MhCalendar;
