CREATE TABLE `announcements` (
  `id`          BIGINT       NOT NULL AUTO_INCREMENT                                          COMMENT '主键ID',
  `admin_id`    BIGINT       NOT NULL                                                         COMMENT '发布管理员ID',
  `title`       VARCHAR(100) NOT NULL                                                         COMMENT '公告标题',
  `content`     TEXT         NOT NULL                                                         COMMENT '公告内容（支持markdown格式）',
  `type`        TINYINT      NOT NULL                                                         COMMENT '公告类型: 0系统公告 1活动公告 2维护通知 3其他',
  `status`      TINYINT      NOT NULL DEFAULT 0                                               COMMENT '状态: 0草稿 1已发布 2已过期',
  `is_top`      TINYINT      NOT NULL DEFAULT 0                                               COMMENT '是否置顶: 0否 1是',
  `view_count`  INT          NOT NULL DEFAULT 0                                               COMMENT '浏览次数',
  `published_at`DATETIME     DEFAULT NULL                                                     COMMENT '发布时间',
  `created_at`  DATETIME     NOT NULL DEFAULT  CURRENT_TIMESTAMP                              COMMENT '创建时间',
  `updated_at`  DATETIME     NOT NULL DEFAULT  CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP  COMMENT '更新时间',
  `is_deleted`  TINYINT      NOT NULL DEFAULT 0                                               COMMENT '逻辑删除: 0否 1是',
  PRIMARY KEY (`id`),
  KEY `idx_admin_id` (`admin_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_general_ci    COMMENT = '公告表';

