CREATE TABLE `order_sepolia`
(
    `id`                 BIGINT          NOT NULL AUTO_INCREMENT COMMENT '主键',
    `marketplace_id`     INT             NOT NULL DEFAULT 0 COMMENT '市场ID',
    `collection_address` VARCHAR(255)    NOT NULL DEFAULT '' COMMENT '所属Collection合约地址',
    `token_id`           VARCHAR(255)    NOT NULL DEFAULT '' COMMENT 'NFT Token ID',
    `order_id`           VARCHAR(255)    NOT NULL COMMENT '订单唯一ID',
    `order_status`       INT             NOT NULL DEFAULT 0 COMMENT '订单状态(0:Active 1:Inactive 2:Expired 3:Cancelled 4:Filled 5:NeedSign)',
    `event_time`         BIGINT          NOT NULL DEFAULT 0 COMMENT '事件时间(秒级时间戳)',
    `expire_time`        BIGINT          NOT NULL DEFAULT 0 COMMENT '过期时间(秒级时间戳)',
    `currency_address`   VARCHAR(255)    NOT NULL DEFAULT '' COMMENT '支付代币合约地址',
    `price`              DECIMAL(30, 10) NOT NULL DEFAULT 0.0000000000 COMMENT '订单价格',
    `maker`              VARCHAR(255)    NOT NULL DEFAULT '' COMMENT '挂单方地址',
    `taker`              VARCHAR(255)    NOT NULL DEFAULT '' COMMENT '吃单方地址',
    `quantity_remaining` BIGINT          NOT NULL DEFAULT 0 COMMENT '剩余可成交数量',
    `size`               BIGINT          NOT NULL DEFAULT 0 COMMENT '订单原始数量',
    `order_type`         BIGINT          NOT NULL DEFAULT 0 COMMENT '订单类型(1:Listing 2:Offer 3:CollectionBid 4:ItemBid)',
    `salt`               BIGINT          NOT NULL DEFAULT 0 COMMENT '签名盐值',
    `create_time`        BIGINT(20)      NOT NULL COMMENT '创建时间(毫秒级时间戳)',
    `update_time`        BIGINT(20)      NOT NULL COMMENT '更新时间(毫秒级时间戳)',
    PRIMARY KEY (`id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_unicode_ci COMMENT ='Sepolia链订单表';

-- 订单唯一标识，防止重复写入
ALTER TABLE `order_sepolia` ADD UNIQUE INDEX `uk_order_id` (`order_id`);

-- 按合约+Token查订单列表（Collection详情页/Item详情页）
ALTER TABLE `order_sepolia` ADD INDEX `idx_collection_token` (`collection_address`, `token_id`);

-- 按挂单方查"我的卖单/买单"
ALTER TABLE `order_sepolia` ADD INDEX `idx_maker` (`maker`);

-- 按状态筛选活跃订单（撮合引擎/前端展示）
ALTER TABLE `order_sepolia` ADD INDEX `idx_order_status` (`order_status`);

-- 按过期时间清理失效订单（定时任务）
ALTER TABLE `order_sepolia` ADD INDEX `idx_expire_time` (`expire_time`);

-- 按创建时间倒序展示最新订单
ALTER TABLE `order_sepolia` ADD INDEX `idx_create_time` (`create_time`);

-- ----------------------------------------------------------------------------------------------
#测试数据
INSERT INTO `order_sepolia` (
    `marketplace_id`, `collection_address`, `token_id`, `order_id`,
    `order_status`, `event_time`, `expire_time`, `currency_address`,
    `price`, `maker`, `taker`, `quantity_remaining`, `size`,
    `order_type`, `salt`, `create_time`, `update_time`
) VALUES
-- ========== SEP-PUNKS (0xA1B2...) ==========
-- 1. Active Listing：当前最低挂牌价 0.07 ETH → 应被识别为实时地板价
(1, '0xA1B2C3D4E5F6A7B8C9D0E1F2A3B4C5D6E7F8A9B0', '8', 'ord-sep-punk-001',
 0, 1754478600, 1757070600, '0x0000000000000000000000000000000000000000',
 0.0700000000, '0xOwner2222222222222222222222222222222222', '', 1, 1,
 1, 100001, 1754478600000, 1754478600000),

-- 2. Active Listing：较高挂牌价 0.12 ETH
(1, '0xA1B2C3D4E5F6A7B8C9D0E1F2A3B4C5D6E7F8A9B0', '9', 'ord-sep-punk-002',
 0, 1754479000, 1757071000, '0x0000000000000000000000000000000000000000',
 0.1200000000, '0xOwner3333333333333333333333333333333333', '', 1, 1,
 1, 100002, 1754479000000, 1754479000000),

-- 3. Filled Listing：已成交，对应 activity #1 (Buy 0.08)
(1, '0xA1B2C3D4E5F6A7B8C9D0E1F2A3B4C5D6E7F8A9B0', '7', 'ord-sep-punk-003',
 4, 1754478000, 1757070000, '0x0000000000000000000000000000000000000000',
 0.0800000000, '0xOwner2222222222222222222222222222222222', '0xBuyer1111111111111111111111111111111111', 0, 1,
 1, 100003, 1754478000000, 1754478000000),

-- 4. Cancelled Listing：已取消，对应 activity #9
(1, '0xA1B2C3D4E5F6A7B8C9D0E1F2A3B4C5D6E7F8A9B0', '22', 'ord-sep-punk-004',
 3, 1754479000, 1757071000, '0x0000000000000000000000000000000000000000',
 0.1000000000, '0xOwner3333333333333333333333333333333333', '', 0, 1,
 1, 100004, 1754465000000, 1754479000000),

-- 5. Expired Listing：已过期，expire_time < 当前时间
(1, '0xA1B2C3D4E5F6A7B8C9D0E1F2A3B4C5D6E7F8A9B0', '30', 'ord-sep-punk-005',
 2, 1754300000, 1754386400, '0x0000000000000000000000000000000000000000',
 0.0900000000, '0xOwner1111111111111111111111111111111111', '', 1, 1,
 1, 100005, 1754300000000, 1754386400000),

-- ========== OLD-SEPOLIA (0xD4E5...) ==========
-- 6. Active Offer：针对特定Item的出价
(2, '0xD4E5F6A7B8C9D0E1F2A3B4C5D6E7F8A9B0C1D2E3', '500', 'ord-og-offer-001',
 0, 1754479500, 1757071500, '0x0000000000000000000000000000000000000000',
 0.1800000000, '0xBuyer6666666666666666666666666666666666', '', 1, 1,
 4, 200001, 1754479500000, 1754479500000),

-- 7. Active CollectionBid：对整个Collection的批量出价
(2, '0xD4E5F6A7B8C9D0E1F2A3B4C5D6E7F8A9B0C1D2E3', '', 'ord-og-cbid-001',
 0, 1754476000, 1757068000, '0x0000000000000000000000000000000000000000',
 0.1000000000, '0xBuyer3333333333333333333333333333333333', '', 5, 10,
 3, 200002, 1754476000000, 1754476000000),

-- ========== GAME-ITEMS (0xE5F6...) ERC1155 ==========
-- 8. Active Listing：ERC1155 部分数量挂单 (剩余3/10)
(1, '0xE5F6A7B8C9D0E1F2A3B4C5D6E7F8A9B0C1D2E3F4', '1001', 'ord-game-list-001',
 0, 1754476000, 1757068000, '0x0000000000000000000000000000000000000000',
 0.0005000000, '0xOwner5555555555555555555555555555555555', '', 3, 10,
 1, 300001, 1754476000000, 1754476000000),

-- 9. NeedSign：待签名订单，不应参与地板价计算
(1, '0xE5F6A7B8C9D0E1F2A3B4C5D6E7F8A9B0C1D2E3F4', '2001', 'ord-game-pending-001',
 5, 1754479800, 1757071800, '0x0000000000000000000000000000000000000000',
 0.0003000000, '0xOwner6666666666666666666666666666666666', '', 200, 200,
 1, 300002, 1754479800000, 1754479800000);