-- 测试种子数据
-- 管理员账号由测试程序自动初始化（initAdmin），此处无需预置

-- 初始化平台
INSERT INTO sys_platform (name, link) VALUES
  ('platformA', 'https://platformA.maplehaze.com'),
  ('platformB', 'https://platformB.maplehaze.com'),
  ('platformC', 'https://platformC.maplehaze.com');
