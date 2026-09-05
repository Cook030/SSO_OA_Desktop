import { MhConModal, MhConModalCloseButton, MhDivider, MhFlex, MhTabs, MhTag, MhTypography } from "@mh-repo/ui";
import React from "react";
import styles from "./index.module.less";

const { Text } = MhTypography;

// ==================== 类型定义（TS 核心）====================
interface AppBasicInfo {
  appAttribute: string;
  downloadUrl: string;
  packageName: string;
  industry: string;
}

interface AppMultiProduct {
  adSpark: string;
  groMore: string;
}

interface AppLogParams {
  appLog: string;
  appKey: string;
}

interface AppDetail {
  name: string;
  id: string;
  platform: string;
  status: string;
  basicInfo: AppBasicInfo;
  multiProduct: AppMultiProduct;
  appLogParams: AppLogParams;
}

interface DetailField {
  label: string;
  value: React.ReactNode;
  fullWidth?: boolean;
}

const appDetailData: AppDetail = {
  name: "快手拉新T7复投用户",
  id: "ID34534563",
  platform: "Android",
  status: "重要状态",
  basicInfo: {
    appAttribute: "正式",
    downloadUrl: "http://zhushou.360.cn/detail?id=4534756&game_src=&apkid=&bind_engine=&from=old_detail",
    packageName: "com.dmzjsq.manhua",
    industry: "uue社区论坛"
  },
  multiProduct: {
    adSpark: "未启用",
    groMore: "未启用"
  },
  appLogParams: {
    appLog: "45645346",
    appKey: "4636457686545346347634635"
  }
};

const detailSections: Array<{ title: string; fields: DetailField[] }> = [
  {
    title: "基础信息",
    fields: [
      { label: "应用属性", value: appDetailData.basicInfo.appAttribute },
      { label: "下载地址", value: appDetailData.basicInfo.downloadUrl, fullWidth: true },
      { label: "程序包名", value: appDetailData.basicInfo.packageName },
      { label: "行业", value: appDetailData.basicInfo.industry }
    ]
  },
  {
    title: "多产品接入",
    fields: [
      { label: "AdSpark", value: appDetailData.multiProduct.adSpark },
      { label: "GroMore", value: appDetailData.multiProduct.groMore }
    ]
  },
  {
    title: "AppLog参数",
    fields: [
      { label: "APPLog", value: appDetailData.appLogParams.appLog },
      { label: "AppKey", value: appDetailData.appLogParams.appKey, fullWidth: true }
    ]
  }
];

const TablePopUpDesc: React.FC<{
  open: boolean;
  onCancel: () => void;
}> = ({ onCancel, open }) => {
  return (
    <MhConModal
      open={open}
      onCancel={onCancel}
      centered
      width={600}
      maskClosable
      title="应用详情"
      headerExtra={<MhConModalCloseButton onClick={onCancel} />}
    >
      <div className={styles.summary}>
        <MhFlex align="center" justify="space-between" gap={12} className={styles.summaryRow}>
          <MhFlex align="center" gap={12} className={styles.summaryMain}>
            <div className={styles.appThumb} aria-hidden />
            <div className={styles.summaryContent}>
              <MhFlex align="center" gap={8} className={styles.titleRow}>
                <MhTypography.Text className={styles.appName}>{appDetailData.name}</MhTypography.Text>
                <MhTag color="green" className={styles.tag}>
                  标签
                </MhTag>
              </MhFlex>
              <MhFlex align="center" gap={8} className={styles.metaRow}>
                <Text className={styles.metaText}>{appDetailData.id}</Text>
                <span className={styles.metaDivider} aria-hidden />
                <Text className={styles.metaText}>{appDetailData.platform}</Text>
              </MhFlex>
            </div>
          </MhFlex>

          <MhFlex align="center" gap={8} className={styles.status}>
            <span className={styles.statusDot} aria-hidden />
            <Text className={styles.statusText}>{appDetailData.status}</Text>
          </MhFlex>
        </MhFlex>
      </div>

      <MhTabs
        defaultActiveKey="media"
        className={styles.tabs}
        items={[
          {
            key: "media",
            label: "媒体详情",
            children: (
              <div className={styles.body}>
                {detailSections.map((section, index) => (
                  <React.Fragment key={section.title}>
                    <section className={styles.section}>
                      <MhTypography.Text className={styles.sectionTitle}>{section.title}</MhTypography.Text>
                      <div className={styles.grid}>
                        {section.fields.map(field => (
                          <div
                            key={field.label}
                            className={`${styles.item}${field.fullWidth ? ` ${styles.itemFull}` : ""}`}
                          >
                            <Text className={styles.label}>{field.label}</Text>
                            <Text className={styles.value}>{field.value}</Text>
                          </div>
                        ))}
                      </div>
                    </section>

                    {index < detailSections.length - 1 ? <MhDivider className={styles.divider} /> : null}
                  </React.Fragment>
                ))}
              </div>
            )
          },
          {
            key: "other",
            label: "其他信息",
            children: (
              <div className={styles.empty}>
                <Text className={styles.emptyText}>暂无相关信息</Text>
              </div>
            )
          }
        ]}
      />
    </MhConModal>
  );
};

export default TablePopUpDesc;
