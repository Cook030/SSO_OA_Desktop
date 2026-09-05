export type ColumnFixed = "left" | "right" | undefined;

export interface TableColumnPreference {
  visibleColumnKeys: string[];
  columnFixedMap: Record<string, ColumnFixed>;
  columnOrder: string[];
}

/** IndexedDB 配置常量 */
const DB_CONFIG = {
  name: "mh-next-ui-preferences",
  storeName: "table-column-preferences",
  version: 1
} as const;

/** 检查是否在浏览器环境且支持 IndexedDB */
const isIndexedDBAvailable = (): boolean => typeof window !== "undefined" && !!window.indexedDB;

/** 获取数据库实例 */
const getDatabase = async (): Promise<IDBDatabase | null> => {
  if (!isIndexedDBAvailable()) return null;

  return new Promise((resolve, reject) => {
    const request = window.indexedDB.open(DB_CONFIG.name, DB_CONFIG.version);

    request.onupgradeneeded = () => {
      const database = request.result;
      if (!database.objectStoreNames.contains(DB_CONFIG.storeName)) {
        database.createObjectStore(DB_CONFIG.storeName);
      }
    };

    request.onsuccess = () => resolve(request.result);
    request.onerror = () => reject(request.error);
  });
};

/** 包装 IDBRequest 为 Promise */
const runRequest = <T>(request: IDBRequest<T>): Promise<T> =>
  new Promise((resolve, reject) => {
    request.onsuccess = () => resolve(request.result);
    request.onerror = () => reject(request.error);
  });

/** 等待事务完成 */
const awaitTransaction = (transaction: IDBTransaction): Promise<void> =>
  new Promise((resolve, reject) => {
    transaction.oncomplete = () => resolve();
    transaction.onerror = () => reject(transaction.error);
    transaction.onabort = () => reject(transaction.error);
  });

/**
 * 获取表格列偏好设置
 * @param storageKey 存储键名
 * @returns 偏好设置对象，如果不存在则返回 null
 */
export const getTableColumnPreference = async (storageKey: string): Promise<TableColumnPreference | null> => {
  const database = await getDatabase();
  if (!database) return null;

  const transaction = database.transaction(DB_CONFIG.storeName, "readonly");
  const store = transaction.objectStore(DB_CONFIG.storeName);

  try {
    const result = await runRequest(store.get(storageKey));
    return result ? (result as TableColumnPreference) : null;
  } finally {
    database.close();
  }
};

/**
 * 保存表格列偏好设置
 * @param storageKey 存储键名
 * @param value 偏好设置对象
 */
export const setTableColumnPreference = async (storageKey: string, value: TableColumnPreference): Promise<void> => {
  const database = await getDatabase();
  if (!database) return;

  const transaction = database.transaction(DB_CONFIG.storeName, "readwrite");
  const store = transaction.objectStore(DB_CONFIG.storeName);

  try {
    await runRequest(store.put(value, storageKey));
    await awaitTransaction(transaction);
  } finally {
    database.close();
  }
};
