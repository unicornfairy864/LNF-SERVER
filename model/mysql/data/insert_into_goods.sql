-- ============================================================
-- 文件：insert_into_goods.sql
-- 用途：校园失物招领系统 - 积分商城商品初始数据（共 14 条）
-- 表结构见 model/mysql/advanced/goods.sql（created_at/updated_at 有默认值，无需写入）
-- 说明：
--   1. 仅用于【空表 goods 的首次初始化】，由人工执行；脚本不含 DELETE、不含事务
--   2. id / created_at / updated_at 由数据库默认值生成，不在语句中指定
--   3. goods 无 name 唯一键（重名由应用层查重），本文件不可重复执行，
--      重复执行会插入同名商品；如需重跑请先清空或改用 WHERE NOT EXISTS 写法
--   4. 字段约束（对应 shop 模块校验）：name 1-100 字符；price 1~1000000；
--      stock >= 0；image_url <=500 字符，相对路径以 / 开头，展示时拼 BASE_URL
--   5. sort_order 号段规则（段内步长 10，预留扩展空间）：
--      10~99  文创生活    100~199 电子数码    200~299 限量/周边
--      900    示例数据（下架/缺货演示，不进默认列表）
--   6. 覆盖场景：在售（含缺货 stock=0）、已下架（is_deleted=1）、
--      有图 / 无图（image_url NULL）、价格区间筛选（10~120 分）
--   7. 请以 UTF-8 编码保存并执行
-- ============================================================
INSERT INTO `goods` (`name`, `description`, `image_url`, `price`, `stock`, `sort_order`, `is_deleted`) VALUES
-- ---------- 文创生活（橙） ----------
('校园卡挂绳',       '定制编织挂绳，蓝白校配色，防丢卡套设计，含透明卡位与快拆扣。', '/uploads/goods/kazhan_lanyard.jpg',  20, 50,  10, 0),
('帆布收纳袋',       '加厚棉麻帆布袋，A4 大小，可装书本水杯，印失物招领主题插画。',   '/uploads/goods/canvas_bag.jpg',      30, 40,  20, 0),
('校园文创笔记本',   'A5 空白内页线装本，封面烫印校徽，120g 米白纸不透墨。',         '/uploads/goods/notebook.jpg',        15, 60,  30, 0),
('中性笔套装',       '0.5mm 黑色按动中性笔 3 支装，速干大容量笔芯。',               NULL,                                10, 80,  40, 0),
('晴雨两用伞',       '三折自动伞，双层黑胶防晒，抗风伞骨，伞面印校园地标剪影。',     '/uploads/goods/umbrella.jpg',        50, 25,  50, 0),
('保温杯',           '316 不锈钢内胆 500ml，长效保温，杯身可刻字（兑换后联系管理员）。', '/uploads/goods/thermos.jpg',    60, 20,  60, 0),
('桌面小风扇',       'USB 台式静音小风扇，三档风速，可夹可立，宿舍自习两用。',       NULL,                                45, 30,  70, 0),
('键盘清洁套装',     '软毛刷 + 粘尘胶 + 清洁泥，机械键盘缝隙除尘专用。',             NULL,                                12, 70,  80, 0),
('手机桌面支架',     '铝合金折叠支架，角度可调，兼容 4-7 英寸手机与平板。',           '/uploads/goods/phone_stand.jpg',     18, 45,  90, 0),
-- ---------- 电子数码（蓝） ----------
('蓝牙耳机收纳盒',   '硬壳防压收纳盒，独立理线仓，适配主流真无线耳机。',             NULL,                                35, 35, 100, 0),
('迷你充电宝 10000mAh', '轻薄移动电源，双向快充，自带 Type-C 线，登机可用。',         '/uploads/goods/power_bank.jpg',     120,  5, 110, 0),
-- ---------- 限量 / 周边（紫） ----------
('工大吉祥物毛绒挂件', '校庆限定毛绒挂件，约 8cm，钥匙扣款式，数量有限。',           '/uploads/goods/mascot_plush.jpg',    80, 15, 200, 0),
('金属徽章套装',       '校徽 + 地标建筑徽章 4 枚装，珐琅工艺，背扣固定。',           NULL,                                25,  0, 210, 0),  -- stock=0：缺货演示（在售但不可兑换）
-- ---------- 示例数据（下架演示，不进公开列表） ----------
('旧版文化衫',         '往届活动文化衫，M 码，已停止发放，仅保留下架示例。',         '/uploads/goods/old_tshirt.jpg',      40, 10, 900, 1); -- is_deleted=1：下架演示

-- ============================================================
-- 执行后自查（可选，直接复制执行）
-- SELECT COUNT(*) FROM goods;                                                      -- 期望 14
-- SELECT COUNT(*) FROM goods WHERE is_deleted = 0;                                 -- 期望 13（公开列表可见）
-- SELECT COUNT(*) FROM goods WHERE is_deleted = 0 AND stock = 0;                   -- 期望 1（徽章套装：缺货但可见）
-- SELECT id, name, price, stock, sort_order FROM goods WHERE is_deleted = 0 ORDER BY sort_order ASC, created_at DESC, id DESC;  -- 默认列表排序
-- SELECT id, name, price FROM goods WHERE is_deleted = 0 AND price >= 10 AND price <= 60 ORDER BY price;  -- 积分区间筛选样例
-- ============================================================
