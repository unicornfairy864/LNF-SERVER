CREATE TABLE `notifications` (
    `id`               BIGINT       NOT NULL AUTO_INCREMENT                                        COMMENT '主键ID',
    `admin_id`         BIGINT       NOT NULL                                                       COMMENT '发布管理员ID',
    `user_id`          BIGINT       NOT NULL                                                       COMMENT '接收用户ID',
    -- 类型扩展（2026-09-30 shop 兑换触发点已接入；纯注释，无表结构变更）：
    -- 5) 积分变动：shop 兑换扣分成功后写入；6) 商品兑换：兑换成功后写入发货提醒    `type`             TINYINT      NOT NULL                                                       COMMENT '类型: 0系统通知 1物品匹配 2认领申请 3认领结果 4评论回复 5积分变动 6商品兑换',
    `title`            VARCHAR(100) NOT NULL                                                       COMMENT '通知标题',
    `content`          TEXT         DEFAULT NULL                                                   COMMENT '通知内容（支持markdown格式）',
    `is_read`          TINYINT      NOT NULL DEFAULT 0                                             COMMENT '是否已读: 0未读 1已读',
    `read_at`          DATETIME     DEFAULT NULL                                                   COMMENT '阅读时间',
    `created_at`       DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP                             COMMENT '创建时间',
    `updated_at`       DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `is_deleted`       TINYINT      NOT NULL DEFAULT 0                                             COMMENT '逻辑删除: 0否 1是',
    PRIMARY KEY (`id`),
-- UnreadCount: WHERE user_id=? AND is_read=0 AND is_deleted=0
    KEY `idx_user_id_read` (`user_id`, `is_read`),
-- List: WHERE user_id=? AND is_deleted=0 [AND is_read/type/admin_id] ORDER BY created_at DESC
-- （type/admin_id 恒与 user_id 复合出现，不设独立索引；管理端独立列表接口落地时再补 admin_id 前缀索引）
    KEY `idx_user_id_created` (`user_id`, `created_at`)
) ENGINE = InnoDB
DEFAULT CHARSET = utf8mb4
COLLATE = utf8mb4_general_ci
COMMENT = '消息通知表';
