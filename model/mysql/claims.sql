CREATE TABLE `claims` (
    `id`               BIGINT       NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `item_id`          BIGINT       NOT NULL                COMMENT '物品ID',
    `claimer_id`       BIGINT       NOT NULL                COMMENT '认领人ID',
    `status`           TINYINT      NOT NULL DEFAULT 0      COMMENT '状态: 0待审核 1已通过 2已拒绝 3已取消',
    `description`      TEXT         DEFAULT NULL            COMMENT '认领描述（支持markdown格式，可嵌入图片链接等）',
    `contact`          VARCHAR(100) DEFAULT NULL            COMMENT '联系方式',
    `auditor_id`       BIGINT       DEFAULT NULL            COMMENT '审核人ID',
    `audit_comment`    VARCHAR(255) DEFAULT NULL            COMMENT '审核意见',
    `audited_at`       DATETIME     DEFAULT NULL            COMMENT '审核时间',
    `created_at`       DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`       DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `is_deleted`       TINYINT      NOT NULL DEFAULT 0      COMMENT '逻辑删除: 0否 1是',
    PRIMARY KEY (`id`),
    KEY `idx_item_id` (`item_id`),
    KEY `idx_claimer_id` (`claimer_id`),
    KEY `idx_auditor_id` (`auditor_id`),
    KEY `idx_status` (`status`),
    KEY `idx_created_at` (`created_at`)
) ENGINE = InnoDB
DEFAULT CHARSET = utf8mb4
COLLATE = utf8mb4_general_ci
COMMENT = '认领记录表';