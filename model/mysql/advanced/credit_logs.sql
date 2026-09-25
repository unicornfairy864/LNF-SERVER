CREATE TABLE `credit_logs` (
    `id`               BIGINT       NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `user_id`          BIGINT       NOT NULL                COMMENT '用户ID',
    `change_amount`    INT          NOT NULL                COMMENT '变动金额（正数为增加，负数为减少）',
    `before_amount`    INT          NOT NULL                COMMENT '变动前积分',
    `after_amount`     INT          NOT NULL                COMMENT '变动后积分',
    `type`             TINYINT      NOT NULL                COMMENT '类型: 0拾金不昧奖励 1认领成功奖励 2违规扣分 3系统调整',
    `related_id`       BIGINT       DEFAULT NULL            COMMENT '关联ID（物品ID/认领ID/举报ID等）',
    `description`      VARCHAR(255) DEFAULT NULL            COMMENT '变动说明',
    `operator_id`      BIGINT       DEFAULT NULL            COMMENT '操作人ID（系统调整时为管理员ID）',
    `created_at`       DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `is_deleted`       TINYINT      NOT NULL DEFAULT 0      COMMENT '逻辑删除: 0否 1是',
    PRIMARY KEY (`id`),
    KEY `idx_user_created` (`user_id`, `created_at`)
) ENGINE = InnoDB
DEFAULT CHARSET = utf8mb4
COLLATE = utf8mb4_general_ci
COMMENT = '积分变动记录表';