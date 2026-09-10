-- =============================================
-- 校园失物招领系统 - 用户表 (user)
-- 数据库: MySQL 5.7+ / 8.0+
-- 字符集: utf8mb4
-- =============================================

CREATE TABLE `user` (
    `id`               BIGINT       NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `username`         VARCHAR(50)  NOT NULL                COMMENT '登录账号(学号/工号)',
    `password_hash`    VARCHAR(100) NOT NULL                COMMENT '加密密码',
    `real_name`        VARCHAR(50)  NOT NULL                COMMENT '真实姓名',
    `qq`               VARCHAR(50)  NOT NULL                COMMENT 'qq号',
    `avatar`           VARCHAR(255) DEFAULT NULL            COMMENT '头像URL',
    `role`             TINYINT      NOT NULL DEFAULT 0      COMMENT '角色: 0普通用户 1失物招领管理员 2系统管理员',
    `status`           TINYINT      NOT NULL DEFAULT 1      COMMENT '状态: 0禁用 1正常',
    `credit`           INT          NOT NULL DEFAULT 0      COMMENT '信誉积分(拾金不昧奖励)',
    `last_login_time`  DATETIME     DEFAULT NULL            COMMENT '最后登录时间',
    `create_time`      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `update_time`      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `is_deleted`       TINYINT      NOT NULL DEFAULT 0      COMMENT '逻辑删除: 0否 1是',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_username` (`username`),
    UNIQUE KEY `uk_qq` (`qq`),
    UNIQUE KEY `uk_user_no` (`user_no`),
    UNIQUE KEY `uk_openid` (`openid`),
    KEY `idx_role` (`role`),
    KEY `idx_status` (`status`),
    KEY `idx_create_time` (`create_time`)
) ENGINE = InnoDB
DEFAULT CHARSET = utf8mb4
COLLATE = utf8mb4_general_ci
COMMENT = '用户表';