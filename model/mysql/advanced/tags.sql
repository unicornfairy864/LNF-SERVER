CREATE TABLE `tags` (
    `id`               BIGINT       NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `name`             VARCHAR(50)  NOT NULL                COMMENT '标签名称',
    `color`            VARCHAR(20)  DEFAULT NULL            COMMENT '标签颜色（十六进制，如#FF5733）',
    `sort_order`       INT          NOT NULL DEFAULT 0      COMMENT '排序序号',
    `created_at`       DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`       DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_name` (`name`),
    KEY `idx_sort_order` (`sort_order`)
) ENGINE = InnoDB
DEFAULT CHARSET = utf8mb4
COLLATE = utf8mb4_general_ci
COMMENT = '标签表';