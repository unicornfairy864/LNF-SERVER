CREATE TABLE `reports` (
    `id`               BIGINT       NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `reporter_id`      BIGINT       NOT NULL                COMMENT '举报人ID',
    `target_type`      TINYINT      NOT NULL                COMMENT '举报对象类型: 0物品 1评论 2用户',
    `target_id`        BIGINT       NOT NULL                COMMENT '举报对象ID',
    `reason`           TINYINT      NOT NULL                COMMENT '举报原因: 0虚假信息 1违规内容 2恶意行为 3其他',
    `description`      TEXT         DEFAULT NULL            COMMENT '举报描述（支持markdown格式）',
    `status`           TINYINT      NOT NULL DEFAULT 0      COMMENT '状态: 0待审核 1已通过 2已拒绝 3已处理',
    `auditor_id`       BIGINT       DEFAULT NULL            COMMENT '审核人ID',
    `audit_comment`    VARCHAR(255) DEFAULT NULL            COMMENT '审核意见',
    `audited_at`       DATETIME     DEFAULT NULL            COMMENT '审核时间',
    `created_at`       DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`       DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `is_deleted`       TINYINT      NOT NULL DEFAULT 0      COMMENT '逻辑删除: 0否 1是',
    PRIMARY KEY (`id`),
    KEY `idx_reporter_id` (`reporter_id`),
    KEY `idx_target_type_id` (`target_type`, `target_id`),
    KEY `idx_auditor_id` (`auditor_id`),
    KEY `idx_status` (`status`),
    KEY `idx_created_at` (`created_at`)
) ENGINE = InnoDB
DEFAULT CHARSET = utf8mb4
COLLATE = utf8mb4_general_ci
COMMENT = '举报表';