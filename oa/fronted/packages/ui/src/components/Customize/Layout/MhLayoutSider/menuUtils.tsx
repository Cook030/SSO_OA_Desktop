import type { MenuProps } from "antd";
import { ExportOutlined } from "../../../General/Icon";
import type { MatchedMenuState, MhLayoutSiderMenuItem } from "./types";

export const normalizePath = (path: string): string => {
  if (!path) return "/";
  if (path === "/") return "/";
  return path.replace(/\/+$/, "") || "/";
};

const isPathLike = (value: string): boolean => value.startsWith("/");

export const matchPath = (pathname: string, target: string): boolean => {
  const normalizedPath = normalizePath(pathname);
  const normalizedTarget = normalizePath(target);
  return normalizedPath === normalizedTarget || normalizedPath.startsWith(`${normalizedTarget}/`);
};

export const findMatchedMenuState = (items: MhLayoutSiderMenuItem[], pathname: string): MatchedMenuState => {
  let best:
    | {
        selectedKey: string;
        ancestorKeys: string[];
        score: number;
      }
    | undefined;

  const visit = (nodes: MhLayoutSiderMenuItem[], ancestors: string[]) => {
    nodes.forEach(node => {
      const candidates: string[] = [];
      if (node.path) candidates.push(node.path);
      if (isPathLike(node.key)) candidates.push(node.key);

      const matchedCandidate = candidates
        .filter(Boolean)
        .sort((a, b) => b.length - a.length)
        .find(candidate => matchPath(pathname, candidate));

      if (matchedCandidate) {
        const score = matchedCandidate.length * 10 + ancestors.length;
        if (!best || score > best.score) {
          best = {
            selectedKey: node.key,
            ancestorKeys: ancestors,
            score
          };
        }
      }

      if (node.children?.length) {
        visit(node.children, [...ancestors, node.key]);
      }
    });
  };

  visit(items, []);

  return {
    selectedKey: best?.selectedKey ?? null,
    ancestorKeys: best?.ancestorKeys ?? []
  };
};

type ConvertMenuContext = {
  onNavigate: (path: string) => void;
  appBasePath?: string;
  styles: {
    menuItemLabel: string;
    menuItemText: string;
    menuItemLink: string;
  };
};

export const convertToAntdMenuItems = (
  items: MhLayoutSiderMenuItem[],
  context: ConvertMenuContext
): NonNullable<MenuProps["items"]> => {
  const { onNavigate, appBasePath, styles } = context;

  return items.map(item => {
    const openInNewTab = Boolean(item.openInNewTab && item.path);
    const menuItem = {
      key: item.key,
      icon: item.icon || null,
      label: openInNewTab ? (
        <div className={styles.menuItemLabel}>
          <span className={styles.menuItemText}>{item.title}</span>
          <span className={styles.menuItemLink} aria-hidden="true">
            <ExportOutlined />
          </span>
        </div>
      ) : (
        item.title
      ),
      children: undefined as NonNullable<MenuProps["items"]> | undefined,
      onClick: undefined as (() => void) | undefined
    };

    if (item.children?.length) {
      menuItem.children = convertToAntdMenuItems(item.children, context);
    } else if (item.path) {
      const path = item.path;
      menuItem.onClick = () => {
        if (openInNewTab) {
          const base = appBasePath?.replace(/\/$/, "") || "";
          window.open(`${base}${path}`, "_blank", "noopener,noreferrer");
          return;
        }
        onNavigate(path);
      };
    }

    return menuItem;
  });
};
