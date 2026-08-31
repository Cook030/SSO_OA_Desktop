import type { MhTreeDataNode } from "@mh-repo/ui";
import { MhButton, MhConModal, MhConModalCloseButton, MhFlex, MhPagination, MhTransfer, MhTree } from "@mh-repo/ui";
import type React from "react";
import type { Key } from "react";
import { useEffect, useMemo, useState } from "react";
import styles from "./index.module.less";

interface TablePopUpMediaAssociationProps {
  open: boolean;
  onCancel: () => void;
  onConfirm?: (selectedKeys: string[]) => void;
}

interface TreeTransferItem {
  key: string;
  title: string;
  children?: TreeTransferItem[];
}

// Mock 树形数据
const mockTreeData: TreeTransferItem[] = [
  {
    key: "media-1",
    title: "媒体1",
    children: [
      { key: "ad-1-1", title: "广告位1-1" },
      { key: "ad-1-2", title: "广告位1-2" },
      { key: "ad-1-3", title: "广告位1-3" },
      { key: "ad-1-4", title: "广告位1-4" }
    ]
  },
  {
    key: "media-2",
    title: "媒体2",
    children: [
      { key: "ad-2-1", title: "广告位2-1" },
      { key: "ad-2-2", title: "广告位2-2" },
      { key: "ad-2-3", title: "广告位2-3" }
    ]
  },
  {
    key: "media-3",
    title: "媒体3",
    children: [
      { key: "ad-3-1", title: "广告位3-1" },
      { key: "ad-3-2", title: "广告位3-2" },
      { key: "ad-3-3", title: "广告位3-3" },
      { key: "ad-3-4", title: "广告位3-4" },
      { key: "ad-3-5", title: "广告位3-5" }
    ]
  },
  {
    key: "media-4",
    title: "媒体4",
    children: [
      { key: "ad-4-1", title: "广告位4-1" },
      { key: "ad-4-2", title: "广告位4-2" }
    ]
  },
  {
    key: "media-5",
    title: "媒体5",
    children: [
      { key: "ad-5-1", title: "广告位5-1" },
      { key: "ad-5-2", title: "广告位5-2" },
      { key: "ad-5-3", title: "广告位5-3" }
    ]
  },
  {
    key: "media-6",
    title: "媒体6",
    children: [
      { key: "ad-6-1", title: "广告位6-1" },
      { key: "ad-6-2", title: "广告位6-2" },
      { key: "ad-6-3", title: "广告位6-3" },
      { key: "ad-6-4", title: "广告位6-4" }
    ]
  }
];

const flattenTreeData = (data: TreeTransferItem[]): TreeTransferItem[] => {
  const result: TreeTransferItem[] = [];
  const traverse = (nodes: TreeTransferItem[]) => {
    nodes.forEach(node => {
      result.push(node);
      if (node.children) {
        traverse(node.children);
      }
    });
  };
  traverse(data);
  return result;
};

const getLeafKeys = (data: TreeTransferItem[]): string[] => {
  const keys: string[] = [];
  const traverse = (nodes: TreeTransferItem[]) => {
    nodes.forEach(node => {
      if (node.children) {
        traverse(node.children);
      } else {
        keys.push(node.key);
      }
    });
  };
  traverse(data);
  return keys;
};

const getAllParentKeys = (data: TreeTransferItem[]): string[] => {
  const keys: string[] = [];
  const traverse = (nodes: TreeTransferItem[]) => {
    nodes.forEach(node => {
      if (node.children && node.children.length > 0) {
        keys.push(node.key);
        traverse(node.children);
      }
    });
  };
  traverse(data);
  return keys;
};

const TablePopUpMediaAssociation: React.FC<TablePopUpMediaAssociationProps> = ({ open, onCancel, onConfirm }) => {
  const [targetKeys, setTargetKeys] = useState<string[]>([]);
  // 核心：由我们组件本体完全受控选中的 keys 状态
  const [selectedKeys, setSelectedKeys] = useState<string[]>([]);

  const [leftPage, setLeftPage] = useState(1);
  const [rightPage, setRightPage] = useState(1);
  const [leftExpandedKeys, setLeftExpandedKeys] = useState<string[]>([]);
  const [rightExpandedKeys, setRightExpandedKeys] = useState<string[]>([]);
  const [leftSearchValue, setLeftSearchValue] = useState("");
  const [rightSearchValue, setRightSearchValue] = useState("");

  const pageSize = 3;
  const dataSource = useMemo(() => flattenTreeData(mockTreeData), []);

  useEffect(() => {
    if (open) {
      setTargetKeys([]);
      setSelectedKeys([]);
      setLeftPage(1);
      setRightPage(1);
      setLeftExpandedKeys([]);
      setRightExpandedKeys([]);
      setLeftSearchValue("");
      setRightSearchValue("");
    }
  }, [open]);

  const handleChange = (keys: Key[]) => {
    setTargetKeys(keys.map(String));
    setSelectedKeys([]); // 数据左右穿梭转移后，立刻清空中间勾选池
  };

  const handleSelectChange = (sourceSelectedKeys: Key[], targetSelectedKeys: Key[]) => {
    setSelectedKeys([...sourceSelectedKeys, ...targetSelectedKeys].map(String));
  };

  const handleConfirm = () => {
    onConfirm?.(targetKeys);
    onCancel();
  };

  const filterTree = (nodes: TreeTransferItem[], isTarget: boolean, searchValue: string = ""): TreeTransferItem[] => {
    const targetSet = new Set(targetKeys);
    const titleMatchSearch = (title: string): boolean => {
      if (!searchValue) return true;
      return title.toLowerCase().includes(searchValue.toLowerCase());
    };

    return nodes
      .map(node => {
        if (node.children) {
          const allValidChildren = filterTree(node.children, isTarget, "");
          if (allValidChildren.length === 0) return null;
          const filteredChildren = filterTree(node.children, isTarget, searchValue);

          if (searchValue) {
            if (titleMatchSearch(node.title) || filteredChildren.length > 0) {
              return {
                ...node,
                children: titleMatchSearch(node.title) ? allValidChildren : filteredChildren
              };
            }
          } else {
            if (filteredChildren.length > 0) {
              return { ...node, children: filteredChildren };
            }
          }
          return null;
        }
        const isInTarget = targetSet.has(node.key);
        const shouldShow = isTarget ? isInTarget : !isInTarget;
        return shouldShow && titleMatchSearch(node.title) ? node : null;
      })
      .filter(Boolean) as TreeTransferItem[];
  };

  const leftTreeData = useMemo(() => filterTree(mockTreeData, false, leftSearchValue), [targetKeys, leftSearchValue]);
  const rightTreeData = useMemo(() => filterTree(mockTreeData, true, rightSearchValue), [targetKeys, rightSearchValue]);

  const leftPageData = useMemo(() => {
    const start = (leftPage - 1) * pageSize;
    return leftTreeData.slice(start, start + pageSize);
  }, [leftTreeData, leftPage]);

  const rightPageData = useMemo(() => {
    const start = (rightPage - 1) * pageSize;
    return rightTreeData.slice(start, start + pageSize);
  }, [rightTreeData, rightPage]);

  const leftTotal = leftTreeData.length;
  const rightTotal = rightTreeData.length;

  useEffect(() => {
    if (!open) return;
    const leftMaxPage = Math.max(1, Math.ceil(leftTotal / pageSize));
    if (leftPage > leftMaxPage) setLeftPage(leftMaxPage);
    if (leftPageData.length > 0) {
      const newExpandedKeys = getAllParentKeys(leftPageData);
      setLeftExpandedKeys(prev => Array.from(new Set([...prev, ...newExpandedKeys])));
    }
  }, [leftTotal, leftPage, leftPageData, open]);

  useEffect(() => {
    if (!open) return;
    const rightMaxPage = Math.max(1, Math.ceil(rightTotal / pageSize));
    if (rightPage > rightMaxPage) setRightPage(rightMaxPage);
    if (rightPageData.length > 0) {
      const newExpandedKeys = getAllParentKeys(rightPageData);
      setRightExpandedKeys(prev => Array.from(new Set([...prev, ...newExpandedKeys])));
    }
  }, [rightTotal, rightPage, rightPageData, open]);

  return (
    <MhConModal
      title="关联媒体广告位"
      open={open}
      onCancel={onCancel}
      width={1000}
      centered
      maskClosable
      className={styles["media-association-modal"]}
      headerExtra={<MhConModalCloseButton onClick={onCancel} />}
      footer={
        <MhFlex justify="flex-end" gap={8}>
          <MhButton onClick={onCancel}>取消</MhButton>
          <MhButton type="primary" onClick={handleConfirm}>
            确定
          </MhButton>
        </MhFlex>
      }
    >
      <MhTransfer
        dataSource={dataSource}
        targetKeys={targetKeys}
        selectedKeys={selectedKeys} // 强绑定顶层状态
        onChange={handleChange}
        onSelectChange={handleSelectChange}
        render={item => item.title}
        showSearch
        showSelectAll={true}
        filterOption={() => true}
        onSearch={(direction, value) => {
          if (direction === "left") {
            setLeftSearchValue(value);
            setLeftPage(1);
          } else {
            setRightSearchValue(value);
            setRightPage(1);
          }
        }}
        listStyle={{ width: 456, height: 534 }}
        className={styles["media-association-transfer"]}
      >
        {({ direction, filteredItems }) => {
          const isTarget = direction === "right";
          const currentPage = isTarget ? rightPage : leftPage;
          const setCurrentPage = isTarget ? setRightPage : setLeftPage;
          const treeData = (isTarget ? rightPageData : leftPageData) as MhTreeDataNode[];
          const totalItems = isTarget ? rightTotal : leftTotal;

          const expandedKeys = isTarget ? rightExpandedKeys : leftExpandedKeys;
          const setExpandedKeys = isTarget ? setRightExpandedKeys : setLeftExpandedKeys;

          // 获取当前过滤面板下所有的可用叶子节点（广告位）
          const allValidItemKeys = useMemo(() => new Set(filteredItems.map(item => String(item.key))), [filteredItems]);
          const currentPanelLeafKeys = useMemo(
            () => getLeafKeys(isTarget ? rightTreeData : leftTreeData),
            [isTarget, rightTreeData, leftTreeData]
          );

          // 计算当前面板实际被选中的叶子节点 Key（清洗脏数据）
          const leafCheckedKeys = useMemo(() => {
            const leafKeySet = new Set(currentPanelLeafKeys);
            return selectedKeys.filter(key => leafKeySet.has(key));
          }, [selectedKeys, currentPanelLeafKeys]);

          // =================【智能状态同步：全选勾 与 半选实心方块】=================
          const { treeCheckedKeys, halfCheckedKeys } = useMemo(() => {
            const checkedSet = new Set(leafCheckedKeys);
            const checkedKeysResult: string[] = [...leafCheckedKeys];
            const halfCheckedKeysResult: string[] = [];

            treeData.forEach(mediaNode => {
              if (mediaNode.children && mediaNode.children.length > 0) {
                const childKeys = getLeafKeys([mediaNode as TreeTransferItem]);
                const checkedChildren = childKeys.filter(key => checkedSet.has(key));

                if (checkedChildren.length === childKeys.length && childKeys.length > 0) {
                  // 所有子节点都被勾选，媒体显示“全选勾”
                  checkedKeysResult.push(String(mediaNode.key));
                } else if (checkedChildren.length > 0) {
                  // 部分子节点勾选，媒体显示“半选方块”
                  halfCheckedKeysResult.push(String(mediaNode.key));
                }
              }
            });

            return { treeCheckedKeys: checkedKeysResult, halfCheckedKeys: halfCheckedKeysResult };
          }, [treeData, leafCheckedKeys]);

          return (
            <div className={styles["transfer-panel-content"]}>
              <div className={styles["transfer-tree-container"]}>
                <MhTree
                  checkable
                  checkStrictly={true} // 保持严格受控
                  treeData={treeData}
                  checkedKeys={{
                    checked: treeCheckedKeys,
                    halfChecked: halfCheckedKeys
                  }}
                  expandedKeys={expandedKeys}
                  onExpand={keys => setExpandedKeys(keys.map(String))}
                  onCheck={(_checked, info) => {
                    const mediaKey = String(info.node.key);
                    const isChecked = info.checked;

                    // =================【🔥核心重构：完全自定义受控操作】=================
                    if (info.node.children && info.node.children.length > 0) {
                      // 1. 如果点击的是父节点（媒体）
                      const childLeafKeys = getLeafKeys([info.node as TreeTransferItem]);
                      const validChildKeys = childLeafKeys.filter(key => allValidItemKeys.has(key));

                      setSelectedKeys(prev => {
                        const baseSet = new Set(prev);
                        if (isChecked) {
                          // 如果是要勾选，把下面所有合法的广告位和媒体自己都塞进去
                          validChildKeys.forEach(k => baseSet.add(k));
                          baseSet.add(mediaKey);
                        } else {
                          // 如果是取消勾选，把下面所有的广告位和媒体自己统统删掉
                          validChildKeys.forEach(k => baseSet.delete(k));
                          baseSet.delete(mediaKey);
                        }
                        return Array.from(baseSet);
                      });
                    } else {
                      // 2. 如果点击的是子节点（广告位）
                      if (allValidItemKeys.has(mediaKey)) {
                        setSelectedKeys(prev => {
                          const baseSet = new Set(prev);
                          if (isChecked) {
                            baseSet.add(mediaKey);
                          } else {
                            baseSet.delete(mediaKey);
                          }
                          return Array.from(baseSet);
                        });
                      }
                    }
                  }}
                />
              </div>
              <div className={styles["transfer-pagination"]}>
                <MhPagination
                  simple
                  current={currentPage}
                  pageSize={pageSize}
                  total={totalItems}
                  onChange={page => setCurrentPage(page)}
                  size="small"
                />
              </div>
            </div>
          );
        }}
      </MhTransfer>
    </MhConModal>
  );
};

export default TablePopUpMediaAssociation;
