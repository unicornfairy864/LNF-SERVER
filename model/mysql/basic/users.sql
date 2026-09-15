CREATE TABLE `users` (
    `id`               BIGINT       NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `username`         VARCHAR(50)  NOT NULL                COMMENT '登录账号',
    `password_hash`    VARCHAR(255) NOT NULL                COMMENT '加密密码',
    `nickname`         VARCHAR(50)  NOT NULL                COMMENT '昵称',
    `realname`         VARCHAR(50)  DEFAULT NULL            COMMENT '真实姓名',
    `gender`           TINYINT      DEFAULT NULL            COMMENT '性别: 0未知 1男 2女',
    `qq`               VARCHAR(50)  DEFAULT NULL            COMMENT 'QQ号',
    `avatar`           VARCHAR(255) DEFAULT NULL            COMMENT '头像URL',
    `role`             TINYINT      NOT NULL DEFAULT 0      COMMENT '角色: 0普通用户 1失物招领管理员 2系统管理员',
    `status`           TINYINT      NOT NULL DEFAULT 1      COMMENT '状态: 0禁用 1正常',
    `credit`           INT          NOT NULL DEFAULT 0      COMMENT '信誉积分(拾金不昧奖励)',
    `last_login_at`    DATETIME     DEFAULT NULL            COMMENT '最后登录时间',
    `created_at`       DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`       DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `is_deleted`       TINYINT      NOT NULL DEFAULT 0      COMMENT '逻辑删除: 0否 1是',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_username` (`username`),
    UNIQUE KEY `uk_qq` (`qq`),
	KEY `idx_realname` (`realname`),
    KEY `idx_role_status` (`role`, `status`),
    KEY `idx_created_at` (`created_at`)
) ENGINE = InnoDB
DEFAULT CHARSET = utf8mb4
COLLATE = utf8mb4_general_ci
COMMENT = '用户表';