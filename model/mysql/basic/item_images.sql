CREATE TABLE `item_images` (
    `id`               BIGINT       NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `item_id`          BIGINT       NOT NULL                COMMENT '物品ID（逻辑关联items.id，不建外键约束）',
    `image_url`        VARCHAR(500) NOT NULL                COMMENT '图片链接（URL或对象存储路径）',
    `sort_order`       TINYINT      NOT NULL DEFAULT 1      COMMENT '展示顺序: 1封面 2第二张 3第三张（每物品最多3张）',
    `created_at`       DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`       DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_item_sort` (`item_id`, `sort_order`),
    CONSTRAINT `chk_item_images_sort_order` CHECK (`sort_order` BETWEEN 1 AND 3)
) ENGINE = InnoDB
DEFAULT CHARSET = utf8mb4
COLLATE = utf8mb4_general_ci
COMMENT = '物品图片表（每个物品最多3张）';