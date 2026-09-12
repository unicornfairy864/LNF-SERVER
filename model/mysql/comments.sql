CREATE TABLE `comments` (
    `id`               BIGINT       NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `item_id`          BIGINT       NOT NULL                COMMENT '物品ID',
    `user_id`          BIGINT       NOT NULL                COMMENT '评论用户ID',
    `parent_id`        BIGINT       DEFAULT NULL            COMMENT '父评论ID（支持回复评论）',
    `content`          TEXT         NOT NULL                COMMENT '评论内容（支持markdown格式）',
    `status`           TINYINT      NOT NULL DEFAULT 1      COMMENT '状态: 0待审核 1正常 2已隐藏',
    `created_at`       DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`       DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `is_deleted`       TINYINT      NOT NULL DEFAULT 0      COMMENT '逻辑删除: 0否 1是',
    PRIMARY KEY (`id`),
    KEY `idx_item_id` (`item_id`),
    KEY `idx_user_id` (`user_id`),
    KEY `idx_parent_id` (`parent_id`),
    KEY `idx_status` (`status`),
    KEY `idx_created_at` (`created_at`)
) ENGINE = InnoDB
DEFAULT CHARSET = utf8mb4
COLLATE = utf8mb4_general_ci
COMMENT = '评论/反馈表';