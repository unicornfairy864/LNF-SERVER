CREATE TABLE `locations` (
    `id`               BIGINT       NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `name`             VARCHAR(100) NOT NULL                COMMENT '地点名称',
    `parent_id`        BIGINT       DEFAULT NULL            COMMENT '父地点ID（支持层级地点）',
    `type`             TINYINT      NOT NULL DEFAULT 0      COMMENT '类型: 0校区 1建筑 2楼层 3房间',
    `address`          VARCHAR(255) DEFAULT NULL            COMMENT '详细地址',
    `sort_order`       INT          NOT NULL DEFAULT 0      COMMENT '排序序号',
    `status`           TINYINT      NOT NULL DEFAULT 1      COMMENT '状态: 0禁用 1启用',
    `created_at`       DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`       DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `is_deleted`       TINYINT      NOT NULL DEFAULT 0      COMMENT '逻辑删除: 0否 1是',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_name_parent` (`name`, `parent_id`),
    KEY `idx_parent_id` (`parent_id`),
    KEY `idx_type` (`type`),
    KEY `idx_sort_order` (`sort_order`),
    KEY `idx_status` (`status`)
) ENGINE = InnoDB
DEFAULT CHARSET = utf8mb4
COLLATE = utf8mb4_general_ci
COMMENT = '地点/位置表';
