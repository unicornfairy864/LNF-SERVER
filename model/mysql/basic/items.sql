CREATE TABLE `items` (
    `id`               BIGINT       NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `user_id`          BIGINT       NOT NULL                COMMENT '发布用户ID',
    `title`            VARCHAR(100) NOT NULL                COMMENT '物品标题',
    `description`      TEXT         NOT NULL                COMMENT '物品描述',
    `type`             TINYINT      NOT NULL                COMMENT '类型: 0丢失 1拾到',
    `status`           TINYINT      NOT NULL DEFAULT 0      COMMENT '状态: 0待审核 1已发布 2已认领 3已关闭',
    `location_id`      BIGINT       DEFAULT NULL            COMMENT '地点ID',
    `location_detail`  VARCHAR(200) DEFAULT NULL            COMMENT '详细地点描述',
    `lost_found_time`  DATETIME     NOT NULL                COMMENT '丢失/拾到时间',
    `contact`          VARCHAR(100) DEFAULT NULL            COMMENT '联系方式',
    `credit_reward`    INT          DEFAULT 0               COMMENT '积分奖励',
    `view_count`       INT          NOT NULL DEFAULT 0      COMMENT '浏览次数',
    `claim_user_id`    BIGINT       DEFAULT NULL            COMMENT '认领用户ID',
    `claim_time`       DATETIME     DEFAULT NULL            COMMENT '认领时间',
    `created_at`       DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`       DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `is_deleted`       TINYINT      NOT NULL DEFAULT 0      COMMENT '逻辑删除: 0否 1是',
    PRIMARY KEY (`id`),
    KEY `idx_user_id` (`user_id`),
    KEY `idx_type_status` (`type`, `status`),
    KEY `idx_location_id` (`location_id`),
    KEY `idx_created_at` (`created_at`)
) ENGINE = InnoDB
DEFAULT CHARSET = utf8mb4
COLLATE = utf8mb4_general_ci
COMMENT = '物品表';