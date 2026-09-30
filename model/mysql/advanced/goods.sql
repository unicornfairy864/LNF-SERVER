CREATE TABLE `goods` (
    `id`          BIGINT       NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `name`        VARCHAR(100) NOT NULL                COMMENT '商品名称',
    `description` TEXT         DEFAULT NULL            COMMENT '商品描述',
    `image_url`   VARCHAR(500) DEFAULT NULL            COMMENT '商品图片URL（单张）',
    `price`       INT          NOT NULL                COMMENT '兑换所需积分',
    `stock`       INT          NOT NULL DEFAULT 0      COMMENT '库存（0不可兑换）',
    `sort_order`  INT          NOT NULL DEFAULT 0      COMMENT '排序序号',
    `created_at`  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `is_deleted`  TINYINT      NOT NULL DEFAULT 0      COMMENT '逻辑删除(下架): 0否 1是',
    PRIMARY KEY (`id`),
-- 默认列表: WHERE is_deleted=0 ORDER BY sort_order ASC, created_at DESC, id DESC
-- （排序含索引外列与混合方向，仍一次 filesort；商品量级小可忽略，2026-09-29 复核确认方案 A）
    KEY `idx_is_deleted_sort` (`is_deleted`, `sort_order`),
-- 查重: WHERE name=? AND is_deleted=0（两列等值走索引；keyword 为 '%kw%' 双侧通配不走本索引）
    KEY `idx_name` (`name`, `is_deleted`, `sort_order`),
-- 积分区间: WHERE is_deleted=0 AND price>=? AND price<=?（等值在前范围居中，行集缩小后小额排序）
    KEY `idx_price` (`is_deleted`, `price`, `sort_order`)
) ENGINE = InnoDB
DEFAULT CHARSET = utf8mb4
COLLATE = utf8mb4_general_ci
COMMENT = '积分商城商品表';
