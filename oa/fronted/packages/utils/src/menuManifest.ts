export interface FlatRouteMenuDefinition {
  key: string;
  title?: string;
  icon?: string;
  path?: string;
  openInNewTab?: boolean;
  routePath?: string;
  index?: boolean;
  parentKey?: string;
  hideInMenu?: boolean;
  group?: boolean;
}

export interface MenuTreeNode {
  title: string;
  key: string;
  icon?: string;
  path?: string;
  children?: MenuTreeNode[];
}

export interface MenuLikeNode {
  title: unknown;
  key: unknown;
  icon?: unknown;
  path?: string;
  children?: MenuLikeNode[];
}

export interface ConfiguredRouteTreeNode {
  key: string;
  index?: boolean;
  path?: string;
  hideInMenu?: boolean;
  children?: ConfiguredRouteTreeNode[];
}

const cloneNode = (node: MenuTreeNode): MenuTreeNode => ({
  ...node,
  children: node.children?.map(cloneNode)
});

export const buildMenuNodesFromFlatRoutes = (definitions: FlatRouteMenuDefinition[]): MenuTreeNode[] => {
  const visibleDefinitions = definitions.filter(definition => !definition.hideInMenu && definition.title);
  const nodeMap = new Map<string, MenuTreeNode>();
  const rootNodes: MenuTreeNode[] = [];

  visibleDefinitions.forEach(definition => {
    nodeMap.set(definition.key, {
      title: definition.title || definition.key,
      key: definition.key,
      icon: definition.icon,
      path: definition.path,
      children: []
    });
  });

  visibleDefinitions.forEach(definition => {
    const currentNode = nodeMap.get(definition.key);
    if (!currentNode) {
      return;
    }

    if (definition.parentKey) {
      const parentNode = nodeMap.get(definition.parentKey);
      if (parentNode) {
        parentNode.children = parentNode.children || [];
        parentNode.children.push(currentNode);
        return;
      }
    }

    rootNodes.push(currentNode);
  });

  const pruneEmptyChildren = (node: MenuTreeNode): MenuTreeNode => {
    const nextNode = cloneNode(node);
    if (nextNode.children?.length) {
      nextNode.children = nextNode.children.map(pruneEmptyChildren);
    } else {
      delete nextNode.children;
    }
    return nextNode;
  };

  return rootNodes.map(pruneEmptyChildren);
};

export const buildMenuTreeFromFlatRoutes = (
  platformKey: string,
  definitions: FlatRouteMenuDefinition[]
): MenuTreeNode[] => {
  return [
    {
      title: "导航菜单",
      key: `${platformKey}-nav`,
      children: buildMenuNodesFromFlatRoutes(definitions)
    }
  ];
};

const cloneMenuLikeNode = <T extends MenuLikeNode>(node: T): T => {
  return {
    ...node,
    children: node.children?.map(child => cloneMenuLikeNode(child as T))
  };
};

const collectMenuNodeMap = <T extends MenuLikeNode>(nodes: T[]): Map<string, T> => {
  const nodeMap = new Map<string, T>();

  const visit = (items: T[]) => {
    items.forEach(item => {
      nodeMap.set(String(item.key), item);
      if (item.children?.length) {
        visit(item.children as T[]);
      }
    });
  };

  visit(nodes);
  return nodeMap;
};

const cloneRemainingMenuNodes = <T extends MenuLikeNode>(nodes: T[], consumedKeys: Set<string>): T[] => {
  return nodes.reduce<T[]>((acc, node) => {
    const nodeKey = String(node.key);
    if (consumedKeys.has(nodeKey)) {
      return acc;
    }

    const clonedNode = cloneMenuLikeNode(node);
    const remainingChildren = clonedNode.children
      ? cloneRemainingMenuNodes(clonedNode.children as T[], consumedKeys)
      : undefined;

    if (remainingChildren?.length) {
      clonedNode.children = remainingChildren;
    } else {
      delete clonedNode.children;
    }

    acc.push(clonedNode);
    return acc;
  }, []);
};

export const mergeMenuNodesWithServerOverlay = <T extends MenuLikeNode>(
  localNodes: T[],
  serverNodes?: Array<Partial<T>> | null
): T[] => {
  if (!serverNodes?.length) {
    return localNodes.map(localNode => cloneMenuLikeNode(localNode));
  }

  const localMap = collectMenuNodeMap(localNodes);
  const consumedKeys = new Set<string>();

  const mergeServerNode = (serverNode: Partial<T>): T | null => {
    const nodeKey = String(serverNode.key || "");
    const localNode = localMap.get(nodeKey);

    if (!localNode || consumedKeys.has(nodeKey)) {
      return null;
    }

    consumedKeys.add(nodeKey);
    const mergedNode = cloneMenuLikeNode(localNode);

    if (typeof serverNode.title === "string" && serverNode.title.trim()) {
      mergedNode.title = serverNode.title as T["title"];
    }

    const localChildren = (localNode.children || []) as T[];
    const serverChildren = (serverNode.children || []) as Array<Partial<T>>;

    if (serverChildren.length > 0) {
      const mergedChildren = serverChildren.map(childNode => mergeServerNode(childNode)).filter(Boolean) as T[];
      const mergedChildKeys = new Set(mergedChildren.map(child => String(child.key)));
      const remainingLocalChildren = localChildren
        .filter(localChild => {
          const childKey = String(localChild.key);
          return !mergedChildKeys.has(childKey) && !consumedKeys.has(childKey);
        })
        .map(localChild => cloneMenuLikeNode(localChild));
      mergedNode.children = [...mergedChildren, ...remainingLocalChildren];
    } else if (localChildren.length > 0 && !("children" in serverNode)) {
      mergedNode.children = mergeMenuNodesWithServerOverlay(localChildren, serverChildren);
    } else {
      delete mergedNode.children;
    }

    return mergedNode;
  };

  const mergedNodes = serverNodes.map(serverNode => mergeServerNode(serverNode)).filter(Boolean) as T[];
  const remainingNodes = cloneRemainingMenuNodes(localNodes, consumedKeys);

  return [...mergedNodes, ...remainingNodes];
};

export const readMenuConfigFromStorage = <T extends MenuLikeNode>(platformKey: string): T[] | null => {
  if (typeof window === "undefined") {
    return null;
  }

  try {
    const raw = window.localStorage.getItem(`menu-config-${platformKey}`);
    if (!raw) {
      return null;
    }

    const parsed = JSON.parse(raw);
    return Array.isArray(parsed) ? (parsed as T[]) : null;
  } catch {
    return null;
  }
};

export const buildConfiguredMenuTree = (
  platformKey: string,
  definitions: FlatRouteMenuDefinition[],
  serverNodes?: Array<Partial<MenuTreeNode>> | null
): MenuTreeNode[] => {
  const defaultTree = buildMenuNodesFromFlatRoutes(definitions);
  const overlayNodes = serverNodes ?? readMenuConfigFromStorage<MenuTreeNode>(platformKey);
  return mergeMenuNodesWithServerOverlay(defaultTree, overlayNodes);
};

export const buildConfiguredRouteTree = (
  platformKey: string,
  definitions: FlatRouteMenuDefinition[],
  serverNodes?: Array<Partial<MenuTreeNode>> | null
): ConfiguredRouteTreeNode[] => {
  const configuredMenuTree = buildConfiguredMenuTree(platformKey, definitions, serverNodes);
  const definitionMap = new Map(definitions.map(definition => [definition.key, definition]));

  const toRouteTree = (nodes: MenuTreeNode[]): ConfiguredRouteTreeNode[] => {
    return nodes.reduce<ConfiguredRouteTreeNode[]>((acc, node) => {
      const definition = definitionMap.get(node.key);
      const childRoutes = node.children?.length ? toRouteTree(node.children) : [];

      if (!definition || definition.group) {
        if (childRoutes.length > 0) {
          acc.push({
            key: node.key,
            children: childRoutes
          });
        }
        return acc;
      }

      if (childRoutes.length > 0) {
        acc.push({
          key: `${node.key}::__group`,
          children: [
            {
              key: definition.key,
              index: definition.index,
              path: definition.routePath
            },
            ...childRoutes
          ]
        });
        return acc;
      }

      acc.push({
        key: definition.key,
        index: definition.index,
        path: definition.routePath
      });
      return acc;
    }, []);
  };

  const configuredRouteTree = toRouteTree(configuredMenuTree);
  const configuredKeys = new Set<string>();

  const collectConfiguredKeys = (nodes: ConfiguredRouteTreeNode[]) => {
    nodes.forEach(node => {
      configuredKeys.add(node.key);
      if (node.children?.length) {
        collectConfiguredKeys(node.children);
      }
    });
  };

  collectConfiguredKeys(configuredRouteTree);

  const hiddenRootRoutes = definitions.reduce<ConfiguredRouteTreeNode[]>((acc, definition) => {
    if (definition.group || definition.parentKey || configuredKeys.has(definition.key)) {
      return acc;
    }

    if (!definition.hideInMenu && definition.title) {
      return acc;
    }

    acc.push({
      key: definition.key,
      index: definition.index,
      path: definition.routePath
    });

    return acc;
  }, []);

  return [...hiddenRootRoutes, ...configuredRouteTree];
};
