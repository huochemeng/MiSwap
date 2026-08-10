CREATE TABLE `collection_sepolia`
(
    `id`                 BIGINT          NOT NULL AUTO_INCREMENT COMMENT '主键',
    `symbol`             VARCHAR(255)    NOT NULL COMMENT '项目标识',
    `chain_id`           INT             NOT NULL DEFAULT 1 COMMENT '链类型(1:以太坊)',
    `auth`               INT             NOT NULL DEFAULT 0 COMMENT '认证(0:默认未认证 1:认证通过 2:认证不通过)',
    `token_standard`     BIGINT          NOT NULL COMMENT '合约实现标准',
    `name`               VARCHAR(255)    NOT NULL COMMENT '项目名称',
    `creator`            VARCHAR(255)    NOT NULL COMMENT '创建者',
    `address`            VARCHAR(255)    NOT NULL COMMENT '链上合约地址',
    `owner_amount`       BIGINT          NOT NULL DEFAULT 0 COMMENT '拥有item人数',
    `item_amount`        BIGINT          NOT NULL DEFAULT 0 COMMENT '该项目NFT的发行总量',
    `description`        TEXT NULL COMMENT '项目描述',
    `website`            VARCHAR(255) NULL COMMENT '项目官网地址',
    `twitter`            VARCHAR(255) NULL COMMENT '项目twitter地址',
    `discord`            VARCHAR(255) NULL COMMENT '项目discord地址',
    `instagram`          VARCHAR(255) NULL COMMENT '项目instagram地址',
    `floor_price`        DECIMAL(30, 10) NOT NULL DEFAULT 0.0000000000 COMMENT '地板价，最低挂牌价',
    `sale_price`         DECIMAL(30, 10) NOT NULL DEFAULT 0.0000000000 COMMENT '成交价，实际卖出价',
    `volume_total`       DECIMAL(30, 10) NOT NULL DEFAULT 0.0000000000 COMMENT '总交易量',
    `image_uri`          VARCHAR(1024) NULL COMMENT '项目封面图的链接',
    `banner_uri`         VARCHAR(1024) NULL COMMENT 'banner image uri',
    `opensea_ban_scan`   INT             NOT NULL DEFAULT 0 COMMENT 'Opensea扫描状态(0:未扫描 1:扫描过)',
    `is_syncing`         INT             NOT NULL DEFAULT 0 COMMENT '是否正在同步(0:否 1:是)',
    `is_need_refresh`    INT             NOT NULL DEFAULT 0 COMMENT '是否需要刷新(0:否 1:是)',
    `history_sale_sync`  INT             NOT NULL DEFAULT 0 COMMENT '历史成交同步状态',
    `history_overview`   INT             NOT NULL DEFAULT 0 COMMENT '历史成交overview(0:已生成 1:等待生成 2:生成错误)',
    `floor_price_status` INT             NOT NULL DEFAULT 0 COMMENT '地板价状态',
    `create_time`        BIGINT(20)      NOT NULL COMMENT '创建时间(毫秒级时间戳)',
    `update_time`        BIGINT(20)      NOT NULL COMMENT '更新时间(毫秒级时间戳)',
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Sepolia链Collection表';

# 针对业务，可以追加索引
-- 按链上地址精确查询（高频）
ALTER TABLE `collection_sepolia` ADD UNIQUE INDEX `uk_address` (`address`);

-- 按 symbol 查询
ALTER TABLE `collection_sepolia` ADD INDEX `idx_symbol` (`symbol`);

-- 按同步状态筛选待处理记录
ALTER TABLE `collection_sepolia` ADD INDEX `idx_is_syncing` (`is_syncing`);
ALTER TABLE `collection_sepolia` ADD INDEX `idx_is_need_refresh` (`is_need_refresh`);

-- 按创建时间排序/分页
ALTER TABLE `collection_sepolia` ADD INDEX `idx_create_time` (`create_time`);

############################################################################################
# 测试数据
INSERT INTO `collection_sepolia` (
    `symbol`, `chain_id`, `auth`, `token_standard`, `name`, `creator`, `address`,
    `owner_amount`, `item_amount`, `description`, `website`, `twitter`, `discord`,
    `instagram`, `floor_price`, `sale_price`, `volume_total`, `image_uri`, `banner_uri`,
    `opensea_ban_scan`, `is_syncing`, `is_need_refresh`, `history_sale_sync`,
    `history_overview`, `floor_price_status`, `create_time`, `update_time`
) VALUES
-- 1. 正常已认证项目，有交易数据，同步完成
('SEP-PUNKS', 1155, 1, 721, 'Sepolia Punks', '0xCreator1111111111111111111111111111111', '0xA1B2C3D4E5F6A7B8C9D0E1F2A3B4C5D6E7F8A9B0',
 1500, 10000, 'The first punk collection on Sepolia testnet.', 'https://sepoliapunks.test', '@sepoliapunks', 'https://discord.gg/sepoliapunks',
 NULL, 0.0500000000, 0.0800000000, 125.5000000000, 'https://img.test/punks-cover.png', 'https://img.test/punks-banner.png',
 1, 0, 0, 1, 0, 1, 1722931200000, 1722934800000),

-- 2. 未认证新项目，刚创建，无交易，正在首次同步
('NEW-APES', 1155, 0, 721, 'New Sepolia Apes', '0xCreator2222222222222222222222222222222', '0xB2C3D4E5F6A7B8C9D0E1F2A3B4C5D6E7F8A9B0C1',
 0, 5000, 'Upcoming ape collection on Sepolia.', NULL, NULL, NULL,
 NULL, 0.0000000000, 0.0000000000, 0.0000000000, NULL, NULL,
 0, 1, 0, 0, 1, 0, 1722938400000, 1722938400000),

-- 3. 认证不通过项目，有历史交易但被标记风险
('SCAM-TOKEN', 1155, 2, 1155, 'Fake Bored Ape', '0xCreator3333333333333333333333333333333', '0xC3D4E5F6A7B8C9D0E1F2A3B4C5D6E7F8A9B0C1D2',
 200, 1000, 'Suspicious collection flagged by community.', NULL, '@fakeape_scam', NULL,
 NULL, 0.0010000000, 0.0020000000, 0.5000000000, 'https://img.test/scam-cover.png', NULL,
 1, 0, 1, 1, 0, 0, 1722844800000, 1722920000000),

-- 4. 老项目，数据需要刷新，overview生成错误
('OLD-SEPOLIA', 1155, 1, 721, 'OG Sepolia Collectibles', '0xCreator4444444444444444444444444444444', '0xD4E5F6A7B8C9D0E1F2A3B4C5D6E7F8A9B0C1D2E3',
 800, 3000, 'One of the oldest collections on Sepolia.', 'https://og-sepolia.test', '@ogsepolia', 'https://discord.gg/ogsep',
 '@ogsepolia_ig', 0.1200000000, 0.1500000000, 450.7500000000, 'https://img.test/og-cover.png', 'https://img.test/og-banner.png',
 1, 0, 1, 1, 2, 1, 1720000000000, 1722900000000),

-- 5. ERC1155标准项目，地板价状态异常
('GAME-ITEMS', 1155, 1, 1155, 'Sepolia Game Assets', '0xCreator5555555555555555555555555555555', '0xE5F6A7B8C9D0E1F2A3B4C5D6E7F8A9B0C1D2E3F4',
 3000, 50000, 'In-game assets for Sepolia RPG.', 'https://sepoliarpg.test', '@sepoliarpg', 'https://discord.gg/sepoliarpg',
 NULL, 0.0005000000, 0.0000000000, 12.3000000000, 'https://img.test/game-cover.png', NULL,
 0, 0, 0, 0, 0, 2, 1722935000000, 1722935000000);