import { MhCard, MhCol, MhDatePicker, MhRow } from "@mh-repo/ui";
import type { Dayjs } from "dayjs";
import dayjs from "dayjs";
import type React from "react";
import { useState } from "react";
import { useNavigate } from "react-router";
import type { BreadcrumbItem } from "../../../components/Breadcrumb";
import PageHeader from "../../../components/PageHeader";

const { RangePicker } = MhDatePicker;

/** 快捷选项配置 */
const rangePresets: { label: string; value: [Dayjs, Dayjs] }[] = [
  { label: "今天", value: [dayjs().startOf("day"), dayjs().endOf("day")] },
  { label: "昨天", value: [dayjs().subtract(1, "day").startOf("day"), dayjs().subtract(1, "day").endOf("day")] },
  { label: "近3天", value: [dayjs().subtract(2, "day").startOf("day"), dayjs().endOf("day")] },
  { label: "最近7天", value: [dayjs().subtract(6, "day").startOf("day"), dayjs().endOf("day")] },
  { label: "最近15天", value: [dayjs().subtract(14, "day").startOf("day"), dayjs().endOf("day")] },
  { label: "最近30天", value: [dayjs().subtract(29, "day").startOf("day"), dayjs().endOf("day")] },
  { label: "上周", value: [dayjs().subtract(1, "week").startOf("week"), dayjs().subtract(1, "week").endOf("week")] },
  { label: "本月", value: [dayjs().startOf("month"), dayjs().endOf("month")] },
  {
    label: "上月",
    value: [dayjs().subtract(1, "month").startOf("month"), dayjs().subtract(1, "month").endOf("month")]
  },
  { label: "去年", value: [dayjs().subtract(1, "year").startOf("year"), dayjs().subtract(1, "year").endOf("year")] }
];

/**
 * 时间选择器组件页面
 * 展示带快捷选项的 RangePicker 和单日期选择
 */
const TimePickerPage: React.FC = () => {
  const navigate = useNavigate();
  const [rangeValue, setRangeValue] = useState<[Dayjs | null, Dayjs | null] | null>(null);
  const [singleValue, setSingleValue] = useState<Dayjs | null>(null);

  /** 面包屑配置 */
  const breadcrumbItems: BreadcrumbItem[] = [
    { key: "dsp", title: "DSP平台", path: "/", clickable: true },
    { key: "components", title: "表单页", path: "/components", clickable: true },
    { key: "timepicker", title: "时间选择器" }
  ];

  /** 处理面包屑点击 */
  const handleBreadcrumbClick = (item: BreadcrumbItem) => {
    if (item.path) {
      navigate(item.path);
    }
  };

  return (
    <div>
      {/* 页面头部 */}
      <PageHeader
        title="时间选择器"
        breadcrumbItems={breadcrumbItems}
        onBreadcrumbClick={handleBreadcrumbClick}
        showFavorite
      />
      <MhCard className="card_minHeight">
        {/* 内容区域 */}
        <MhRow gutter={[24, 24]}>
          {/* 起止时间 */}
          <MhCol span={8}>
            <div style={{ marginBottom: 16 }}>
              <span style={{ fontSize: 14, color: "rgba(0, 0, 0, 0.88)", fontWeight: 400 }}>起止时间</span>
            </div>
            <RangePicker
              style={{ width: "100%" }}
              placeholder={["开始时间", "结束时间"]}
              presets={rangePresets}
              value={rangeValue}
              onChange={dates => setRangeValue(dates)}
            />
          </MhCol>

          {/* 单时间选择 */}
          <MhCol span={8} offset={6}>
            <div style={{ marginBottom: 16 }}>
              <span style={{ fontSize: 14, color: "rgba(0, 0, 0, 0.88)", fontWeight: 400 }}>单时间选择</span>
            </div>
            <MhDatePicker
              style={{ width: "100%" }}
              placeholder="请选择日期"
              value={singleValue}
              onChange={date => setSingleValue(date as Dayjs | null)}
            />
          </MhCol>

          {/* 起止时间（带时分） */}
          <MhCol span={8}>
            <div style={{ marginBottom: 16 }}>
              <span style={{ fontSize: 14, color: "rgba(0, 0, 0, 0.88)", fontWeight: 400 }}>起止时间（带时分）</span>
            </div>
            <RangePicker
              style={{ width: "100%" }}
              placeholder={["开始时间", "结束时间"]}
              //   presets={rangePresets}
              showTime={{ format: "HH:mm", showSecond: false }}
              format="YYYY-MM-DD HH:mm"
            />
          </MhCol>

          {/* 单时间选择（带时分） */}
          <MhCol span={8} offset={6}>
            <div style={{ marginBottom: 16 }}>
              <span style={{ fontSize: 14, color: "rgba(0, 0, 0, 0.88)", fontWeight: 400 }}>单时间选择（带时分）</span>
            </div>
            <MhDatePicker
              style={{ width: "100%" }}
              placeholder="请选择日期"
              showTime={{ format: "HH:mm", showSecond: false }}
              format="DD/MM/YY HH:mm"
            />
          </MhCol>
        </MhRow>
      </MhCard>
    </div>
  );
};

export default TimePickerPage;
