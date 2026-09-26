CREATE TABLE `notifications` (
    `id`               BIGINT       NOT NULL AUTO_INCREMENT                                        COMMENT '主键ID',
    `admin_id`         BIGINT       NOT NULL                                                       COMMENT '发布管理员ID',
    `user_id`          BIGINT       NOT NULL                                                       COMMENT '接收用户ID',
    `type`             TINYINT      NOT NULL                                                       COMMENT '类型: 0系统通知 1物品匹配 2认领申请 3认领结果 4评论回复 5积分变动',
    `title`            VARCHAR(100) NOT NULL                                                       COMMENT '通知标题',
    `content`          TEXT         DEFAULT NULL                                                   COMMENT '通知内容（支持markdown格式）',
    `is_read`          TINYINT      NOT NULL DEFAULT 0                                             COMMENT '是否已读: 0未读 1已读',
    `read_at`          DATETIME     DEFAULT NULL                                                   COMMENT '阅读时间',
    `created_at`       DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP                             COMMENT '创建时间',
    `updated_at`       DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `is_deleted`       TINYINT      NOT NULL DEFAULT 0                                             COMMENT '逻辑删除: 0否 1是',
    PRIMARY KEY (`id`),
    KEY `idx_type` (`type`),
	KEY `idx_user_id_read` (`user_id`, `is_read`),
    KEY `idx_user_id_created` (`user_id`, `created_at`),
    KEY `idx_admin_id_created_at` (`admin_id`, `created_at`)
) ENGINE = InnoDB
DEFAULT CHARSET = utf8mb4
COLLATE = utf8mb4_general_ci
COMMENT = '消息通知表';
