CREATE TABLE `item_tags` (
    `id`               BIGINT       NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `item_id`          BIGINT       NOT NULL                COMMENT '物品ID',
    `tag_id`           BIGINT       NOT NULL                COMMENT '标签ID',
    `created_at`       DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_item_tag` (`item_id`, `tag_id`),
    KEY `idx_tag_item` (`tag_id`, `item_id`)
) ENGINE = InnoDB
DEFAULT CHARSET = utf8mb4
COLLATE = utf8mb4_general_ci
COMMENT = '物品标签关联表';
