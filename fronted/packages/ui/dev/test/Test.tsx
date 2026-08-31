import { Button, Pagination, Tree } from "antd";
import type { DataNode, TreeProps } from "antd/es/tree";
import type React from "react";
import { useEffect, useState } from "react";

// 模拟树形数据（扁平数组，方便分页）
const mockAllData: DataNode[] = Array.from({ length: 50 }).map((_, i) => ({
  key: `key-${i}`,
  title: `节点 ${i + 1}`,
  children:
    i % 5 === 0
      ? [
          { key: `key-${i}-1`, title: `子节点 ${i + 1}-1` },
          { key: `key-${i}-2`, title: `子节点 ${i + 1}-2` }
        ]
      : undefined
}));

const PAGE_SIZE = 10;

const Test: React.FC = () => {
  // 分页
  const [current, setCurrent] = useState(1);
  const [leftData, setLeftData] = useState<DataNode[]>([]);

  // 选中keys（左树、右树）
  const [leftChecked, setLeftChecked] = useState<string[]>([]);
  const [targetKeys, setTargetKeys] = useState<string[]>([]);
  const [rightData, setRightData] = useState<DataNode[]>([]);

  // 分页切页：截取当前页的树数据
  useEffect(() => {
    const start = (current - 1) * PAGE_SIZE;
    const end = start + PAGE_SIZE;
    setLeftData(mockAllData.slice(start, end));
  }, [current]);

  // 把左树选中的移到右边
  const handleMoveToRight = () => {
    const moveKeys = leftChecked.filter(k => !targetKeys.includes(k));
    setTargetKeys(prev => [...prev, ...moveKeys]);
    // 右树数据简单合并（实际可按key去重）
    const moveNodes = collectNodesByKeys(mockAllData, moveKeys);
    setRightData(prev => [...prev, ...moveNodes]);
    setLeftChecked([]);
  };

  // 把右边选中的移回左边
  const handleMoveToLeft = () => {
    setTargetKeys(prev => prev.filter(k => !leftChecked.includes(k)));
    setRightData(prev => prev.filter(n => !leftChecked.includes(n.key)));
    setLeftChecked([]);
  };

  // 递归收集指定key的节点（含子）
  const collectNodesByKeys = (nodes: DataNode[], keys: string[]): DataNode[] => {
    const res: DataNode[] = [];
    const loop = (ns: DataNode[]) => {
      ns.forEach(n => {
        if (keys.includes(n.key as string)) res.push(n);
        if (n.children) loop(n.children);
      });
    };
    loop(nodes);
    return res;
  };

  const onLeftCheck: TreeProps["onCheck"] = checkedKeys => {
    setLeftChecked(checkedKeys as string[]);
  };

  return (
    <div style={{ display: "flex", alignItems: "flex-start", gap: 8 }}>
      {/* 左侧：Tree + 分页 */}
      <div style={{ width: 250, border: "1px solid #eee", padding: 8 }}>
        <Tree checkable treeData={leftData} checkedKeys={leftChecked} onCheck={onLeftCheck} defaultExpandAll />
        <Pagination
          current={current}
          pageSize={PAGE_SIZE}
          total={mockAllData.length}
          onChange={setCurrent}
          size="small"
          style={{ marginTop: 8, textAlign: "center" }}
        />
      </div>

      {/* 中间按钮 */}
      <div style={{ display: "flex", flexDirection: "column", gap: 8, marginTop: 20 }}>
        <Button type="primary" onClick={handleMoveToRight} disabled={!leftChecked.length}>
          →
        </Button>
        <Button onClick={handleMoveToLeft} disabled={!leftChecked.length}>
          ←
        </Button>
      </div>

      {/* 右侧：已选中Tree */}
      <div style={{ width: 250, border: "1px solid #eee", padding: 8 }}>
        <Tree checkable treeData={rightData} checkedKeys={leftChecked} onCheck={onLeftCheck} defaultExpandAll />
      </div>
    </div>
  );
};

export default Test;
