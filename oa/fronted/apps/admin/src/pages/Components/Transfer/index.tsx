import { MhCard, MhCascader, MhCol, MhForm, MhRow, MhSelect, MhTag, MhTransfer } from "@mh-repo/ui";
import type React from "react";
import { useState } from "react";
import { useNavigate } from "react-router";
import type { BreadcrumbItem } from "../../../components/Breadcrumb";
import PageHeader from "../../../components/PageHeader";

const { Item: FormItem } = MhForm;
const { Option } = MhSelect;

/**
 * 穿梭框组件页面
 * 展示带标签的Transfer、Select、Cascader等组件
 */
const TransferPage: React.FC = () => {
  const navigate = useNavigate();
  const [form] = MhForm.useForm();

  //   // Transfer 数据
  const [targetKeys, setTargetKeys] = useState<string[]>([]);
  const [selectedKeys, setSelectedKeys] = useState<string[]>([]);

  // 图片styleid 选项数据
  const imageStyleOptions = [
    { value: "1", label: "OptionOption", tags: ["标签1", "标签2"] },
    { value: "2", label: "OptionOptionOption", tags: ["标签1", "标签2"] },
    { value: "3", label: "Option", tags: ["标签"] },
    { value: "4", label: "Option", tags: ["标签"] },
    { value: "5", label: "Option", tags: ["标签"] }
  ];

  // 广告位选择 级联数据
  const adPositionOptions = [
    {
      value: "cas432432432c",
      label: "Casc",
      tags: ["标签1", "标签2"],
      children: [
        { value: "casecas23213ecase", label: "Casecasecase", tags: ["标签1", "标签2"] },
        {
          value: "cas1case3213case",
          label: "Cascasecase",
          tags: ["标签2"],
          children: [{ value: "cas1cas23121132343ecase", label: "Cascasecase", tags: ["标签2"] }]
        }
      ]
    },
    {
      value: "cas435435c2",
      label: "Casc2",
      tags: ["标签1", "标签2"],
      children: [
        { value: "casecasecase", label: "Casecasecase", tags: ["标签1", "标签2"] },
        {
          value: "cas1casecas321321e",
          label: "Cascasecase",
          tags: ["标签2"],
          children: [{ value: "cas1casecase", label: "Cascasecase", tags: ["标签2"] }]
        }
      ]
    },
    {
      value: "casc3",
      label: "Casc3",
      tags: ["标签1"],
      children: [
        { value: "casecasecase", label: "Casecasecase", tags: ["标签1"] },
        { value: "casecasecase2", label: "Casecasecase", tags: ["标签2"] },
        {
          value: "cas1casecase3",
          label: "Cascasecase",
          tags: ["标签2"],
          children: [{ value: "cas1casecase", label: "Cascasecase", tags: ["标签2"] }]
        }
      ]
    }
  ];

  // 多级单选 级联数据
  const multiLevelOptions = [
    {
      value: "cascader1",
      label: "Cascader",
      tags: ["标签"],
      children: [
        {
          value: "cascader2",
          label: "Cascader",
          tags: ["标签1", "标签2"],
          children: [{ value: "cascader3", label: "Cascader", tags: ["标签"] }]
        },
        {
          value: "cascader4",
          label: "Cascader",
          tags: ["标签"],
          children: [{ value: "cascader5", label: "Cascader", tags: ["标签1", "标签2"] }]
        }
      ]
    },
    {
      value: "cascader6",
      label: "Cascader",
      tags: ["标签"],
      children: [
        {
          value: "cascader7",
          label: "Cascader",
          tags: ["标签1", "标签2"],
          children: [{ value: "cascader8", label: "Cascader", tags: ["标签1", "标签2"] }]
        }
      ]
    }
  ];

  // Transfer 数据源 - 带标签
  const transferDataSource = [
    { key: "1", title: "Content", tags: ["标签"] },
    { key: "2", title: "Content", tags: ["标签"] },
    { key: "3", title: "Content", tags: ["标签", "标签2"] },
    { key: "4", title: "Content", tags: ["标签"] },
    { key: "5", title: "Content", tags: ["标签"] },
    { key: "6", title: "Content", tags: ["标签"] },
    { key: "7", title: "Content", tags: ["标签", "标签2"] },
    { key: "8", title: "Content", tags: ["标签"] }
  ];

  /** 面包屑配置 */
  const breadcrumbItems: BreadcrumbItem[] = [
    { key: "dsp", title: "DSP平台", path: "/", clickable: true },
    { key: "components", title: "表单页", path: "/components", clickable: true },
    { key: "transfer", title: "多级选择器" }
  ];

  /** 处理面包屑点击 */
  const handleBreadcrumbClick = (item: BreadcrumbItem) => {
    if (item.path) {
      navigate(item.path);
    }
  };

  /** Transfer 渲染每一项 - 带标签 */
  const renderTransferItem = (item: { key: string; title: string; tags?: string[] }) => {
    return {
      label: (
        <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between", width: "100%" }}>
          <span>{item.title}</span>
          <span style={{ display: "flex", gap: 4 }}>
            {item.tags?.slice(0, 2).map((tag, index) => (
              <MhTag key={index} color="green" style={{ margin: 0, fontSize: 12 }}>
                {tag}
              </MhTag>
            ))}
          </span>
        </div>
      ),
      value: item.title
    };
  };

  /** Select 自定义下拉渲染 - 带标签 */
  const renderSelectOption = (option: { value: string; label: string; tags?: string[] }) => (
    <Option key={option.value} value={option.value}>
      <div style={{ display: "flex", alignItems: "center" }}>
        <span>{option.label}</span>
        {option.tags && option.tags.length > 0 && (
          <span style={{ display: "flex", gap: 6, marginRight: 8, marginLeft: 8 }}>
            {option.tags.slice(0, 2).map((tag, index) => (
              <MhTag key={index} color="green" style={{ margin: 0, fontSize: 12 }}>
                {tag}
              </MhTag>
            ))}
          </span>
        )}
      </div>
    </Option>
  );

  /** Cascader 自定义选项渲染 - 带标签 */
  const cascaderTagRender = (option: { label?: React.ReactNode; tags?: string[] }) => {
    return (
      <div style={{ display: "flex", alignItems: "center", justifyContent: "flex-start", width: "100%" }}>
        <span>{option.label}</span>
        {option.tags && option.tags.length > 0 && (
          <span style={{ display: "flex", gap: 4, marginLeft: 8, justifyContent: "flex-start" }}>
            {option.tags.slice(0, 2).map((tag, index) => (
              <MhTag key={index} color="green" style={{ margin: 0, fontSize: 12 }}>
                {tag}
              </MhTag>
            ))}
          </span>
        )}
      </div>
    );
  };

  return (
    <div>
      {/* 页面头部 */}
      <PageHeader
        title="多级选择器"
        breadcrumbItems={breadcrumbItems}
        onBreadcrumbClick={handleBreadcrumbClick}
        showFavorite
      />
      <MhCard className="card_minHeight">
        {/* 表单内容 */}
        <MhForm form={form} layout="vertical">
          <MhRow gutter={[24, 0]}>
            <MhCol span={8}>
              <FormItem label="单选" name="imageStyle">
                <MhSelect placeholder="请选择" style={{ width: "100%" }} dropdownStyle={{ minWidth: 200 }} allowClear>
                  {imageStyleOptions.map(renderSelectOption)}
                </MhSelect>
              </FormItem>
            </MhCol>

            <MhCol span={8} offset={6}>
              <FormItem label="多级多选" name="adPosition">
                <MhCascader
                  placeholder="请选择"
                  style={{ width: "100%" }}
                  options={adPositionOptions}
                  multiple
                  optionRender={cascaderTagRender}
                />
              </FormItem>
            </MhCol>

            {/* 多级单选 */}
            <MhCol span={8}>
              <FormItem label="多级单选" name="multiLevel">
                <MhCascader
                  placeholder="请选择"
                  style={{ width: "100%" }}
                  options={multiLevelOptions}
                  optionRender={cascaderTagRender}
                />
              </FormItem>
            </MhCol>

            {/* 穿梭框 - 带标签 */}
            <MhCol span={24}>
              <FormItem label="穿梭框组件" name="transfer">
                <MhTransfer
                  dataSource={transferDataSource}
                  targetKeys={targetKeys}
                  selectedKeys={selectedKeys}
                  onChange={(nextTargetKeys, direction, moveKeys) => {
                    setTargetKeys(nextTargetKeys as string[]);
                    console.log("targetKeys:", nextTargetKeys);
                    console.log("direction:", direction);
                    console.log("moveKeys:", moveKeys);
                  }}
                  onSelectChange={(sourceSelectedKeys, targetSelectedKeys) => {
                    setSelectedKeys([...(sourceSelectedKeys as string[]), ...(targetSelectedKeys as string[])]);
                  }}
                  render={item => renderTransferItem(item as { key: string; title: string; tags?: string[] }).label}
                  titles={["Source", "Target"]}
                  listStyle={{
                    width: 300,
                    height: 400
                  }}
                />
              </FormItem>
            </MhCol>
          </MhRow>
        </MhForm>
      </MhCard>
    </div>
  );
};

export default TransferPage;
