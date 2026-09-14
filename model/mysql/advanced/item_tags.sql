CREATE TABLE `item_tags` (
    `id`               BIGINT       NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `item_id`          BIGINT       NOT NULL                COMMENT '物品ID',
    `tag_id`           BIGINT       NOT NULL                COMMENT '标签ID',
    `sort_order`       INT          NOT NULL DEFAULT 0      COMMENT '排序序号',
    `created_at`       DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `is_deleted`       TINYINT      NOT NULL DEFAULT 0      COMMENT '逻辑删除: 0否 1是',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_item_tag` (`item_id`, `tag_id`),
    KEY `idx_item_id` (`item_id`),
    KEY `idx_tag_id` (`tag_id`)
) ENGINE = InnoDB
DEFAULT CHARSET = utf8mb4
COLLATE = utf8mb4_general_ci
COMMENT = '物品标签关联表';
