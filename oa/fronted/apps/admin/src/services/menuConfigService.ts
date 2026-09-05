import type { MenuItem } from "@mh-repo/types";
import { GraphqlBizError, sdk } from "../generated/service/useSdk";

/**
 * 菜单配置数据结构（解包后）
 */
interface MenuConfigData {
  platforms?: Array<{ key: string; label: string }>;
  allMenus?: Record<string, any[]>;
}

/**
 * 从树节点转换为菜单项
 */
function convertTreeToMenuItems(treeNodes: any[]): MenuItem[] {
  if (!treeNodes || treeNodes.length === 0) return [];

  const convertNode = (node: any): MenuItem | null => {
    // 跳过根节点"导航菜单"
    if (node.key?.includes("-nav")) {
      if ("children" in node) {
        return {
          children: Array.isArray(node.children) ? node.children.map(convertNode).filter(Boolean) : []
        } as any;
      }
      return null;
    }

    const menuItem: MenuItem = {
      key: node.key || "",
      title: node.title || "",
      icon: node.icon
    };

    if ("children" in node) {
      menuItem.children = Array.isArray(node.children)
        ? (node.children.map(convertNode).filter(Boolean) as MenuItem[])
        : [];
    } else if (node.path) {
      menuItem.path = node.path;
    }

    return menuItem;
  };

  // 处理根节点
  const rootNode = treeNodes[0];
  if (rootNode?.children) {
    return rootNode.children.map(convertNode).filter(Boolean) as MenuItem[];
  }

  return treeNodes.map(convertNode).filter(Boolean) as MenuItem[];
}

/**
 * 从 API 获取菜单配置
 * @param platformKey 平台标识（如 'dsp', 'ssp'）
 * @returns 菜单项数组，失败返回 null
 */
export async function fetchMenuConfig(platformKey: string): Promise<MenuItem[] | null> {
  try {
    // wrapper 已统一解包 {code, msg, data}，这里直接拿到 data
    const result = await sdk.sso2.mhSso2_getMenuConfig_query();
    const config = result.mhSso2_getMenuConfig as MenuConfigData | null | undefined;

    const platformMenu = config?.allMenus?.[platformKey];
    if (!platformMenu) {
      console.warn(`未找到平台 ${platformKey} 的菜单配置`);
      return null;
    }

    // 保存到 localStorage 作为缓存
    localStorage.setItem(`menu-config-${platformKey}`, JSON.stringify(platformMenu));

    return convertTreeToMenuItems(platformMenu);
  } catch (error) {
    if (error instanceof GraphqlBizError) {
      console.warn(`获取菜单配置失败: ${error.msg}`);
    } else {
      console.error("获取菜单配置时出错:", error);
    }
    return null;
  }
}

/**
 * 从 localStorage 加载菜单配置
 * @param platformKey 平台标识
 * @returns 菜单项数组，失败返回 null
 */
export function loadMenuConfigFromCache(platformKey: string): MenuItem[] | null {
  try {
    const savedConfig = localStorage.getItem(`menu-config-${platformKey}`);
    if (!savedConfig) return null;

    const treeData = JSON.parse(savedConfig);
    const convertedMenu = convertTreeToMenuItems(treeData);

    return convertedMenu.length > 0 ? convertedMenu : null;
  } catch (error) {
    console.error("从缓存加载菜单配置失败:", error);
    return null;
  }
}
