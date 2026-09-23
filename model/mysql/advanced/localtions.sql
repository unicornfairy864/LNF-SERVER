CREATE TABLE `locations` (
    `id`               BIGINT       NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `name`             VARCHAR(100) NOT NULL                COMMENT '地点名称',
    `parent_id`        BIGINT       NOT NULL                COMMENT '父地点ID（支持层级地点）',
    `level`            INT          NOT NULL                COMMENT '树深度',
    `address`          VARCHAR(255) DEFAULT NULL            COMMENT '详细地址',
    `sort_order`       INT          NOT NULL DEFAULT 0      COMMENT '排序序号',
    `created_at`       DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`       DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    KEY `idx_parent_sort` (`parent_id`, `sort_order`),
    KEY `idx_level_sort` (`level`, `sort_order`),
    KEY `idx_name` (`name`)
) ENGINE = InnoDB
DEFAULT CHARSET = utf8mb4
COLLATE = utf8mb4_general_ci
COMMENT = '地点/位置表';
