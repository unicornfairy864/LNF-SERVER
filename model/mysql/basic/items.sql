CREATE TABLE `items` (
    `id`               BIGINT       NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `user_id`          BIGINT       NOT NULL                COMMENT '发布用户ID',
    `title`            VARCHAR(100) NOT NULL                COMMENT '物品标题',
    `description`      TEXT         NOT NULL                COMMENT '物品描述',
    `type`             TINYINT      NOT NULL                COMMENT '类型: 0丢失 1拾到',
    `status`           TINYINT      NOT NULL DEFAULT 0      COMMENT '状态: 0已发布 1已认领 2已关闭',
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
-- 首页信息流: WHERE is_deleted=0 AND type=? AND status=0 ORDER BY lost_found_time DESC
    KEY `idx_items_home` (`is_deleted`, `type`, `status`, `lost_found_time`),
-- 我的发布(默认全部状态): WHERE user_id=? AND is_deleted=0 ORDER BY created_at DESC, id DESC
    KEY `idx_items_user` (`user_id`, `is_deleted`, `created_at`),
-- 状态筛选+时间排序: WHERE is_deleted=0 AND status=0 ORDER BY created_at
    KEY `idx_items_audit` (`is_deleted`, `status`, `created_at`),
-- 信息流×地点筛选: WHERE location_id=? AND is_deleted=0 AND type=? AND status=0 ORDER BY lost_found_time DESC
    KEY `idx_items_location` (`location_id`, `is_deleted`, `type`, `status`, `lost_found_time`),
-- 我的认领: WHERE claim_user_id=? AND is_deleted=0 ORDER BY claim_time DESC
    KEY `idx_items_claim_user` (`claim_user_id`, `is_deleted`, `claim_time`),
-- 关键词搜索(覆盖描述、包含语义): MATCH(title, description) AGAINST(? IN BOOLEAN MODE)
    FULLTEXT KEY `ft_items_search` (`title`, `description`) WITH PARSER ngram
) ENGINE = InnoDB
DEFAULT CHARSET = utf8mb4
COLLATE = utf8mb4_general_ci
COMMENT = '物品表';