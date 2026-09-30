CREATE TABLE `orders` (
    `id`         BIGINT      NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `order_no`   VARCHAR(32) NOT NULL                COMMENT '业务单号（时间戳+随机，幂等展示用）',
    `user_id`    BIGINT      NOT NULL                COMMENT '兑换用户ID',
    `goods_id`   BIGINT      NOT NULL                COMMENT '商品ID',
    `goods_name` VARCHAR(100) NOT NULL               COMMENT '商品名称快照（防商品后续改名/删除）',
    `price`      INT         NOT NULL                COMMENT '成交积分快照',
    `qq`         VARCHAR(50) NOT NULL                COMMENT '用户QQ快照（发群消息at用）',
    `nickname`   VARCHAR(50) NOT NULL                COMMENT '用户昵称快照（发群消息用）',
    `created_at` DATETIME    NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '兑换时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_order_no` (`order_no`),
    KEY `idx_user_created` (`user_id`, `created_at`),
    KEY `idx_goods_created` (`goods_id`, `created_at`)
) ENGINE = InnoDB
DEFAULT CHARSET = utf8mb4
COLLATE = utf8mb4_general_ci
COMMENT = '积分兑换订单表';
