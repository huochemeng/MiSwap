CREATE TABLE `item_sepolia`
(
    `id`                 BIGINT          NOT NULL AUTO_INCREMENT COMMENT '主键',
    `chain_id`           INT             NOT NULL DEFAULT 1 COMMENT '链类型',
    `collection_address` VARCHAR(255)    NOT NULL DEFAULT '' COMMENT '所属Collection合约地址',
    `token_id`           VARCHAR(255)    NOT NULL COMMENT 'NFT Token ID',
    `name`               VARCHAR(255)    NOT NULL COMMENT 'NFT名称',
    `owner`              VARCHAR(255)    NOT NULL DEFAULT '' COMMENT '当前拥有者地址',
    `creator`            VARCHAR(255)    NOT NULL COMMENT '创建者地址',
    `supply`             BIGINT          NOT NULL DEFAULT 0 COMMENT '最大发行份数',
    `list_price`         DECIMAL(30, 10) NOT NULL DEFAULT 0.0000000000 COMMENT '上架价格',
    `list_time`          BIGINT          NOT NULL DEFAULT 0 COMMENT '上架时间(毫秒级时间戳)',
    `sale_price`         DECIMAL(30, 10) NOT NULL DEFAULT 0.0000000000 COMMENT '销售价格',
    `views`              BIGINT          NOT NULL DEFAULT 0 COMMENT '浏览量',
    `create_time`        BIGINT(20)      NOT NULL COMMENT '创建时间(毫秒级时间戳)',
    `update_time`        BIGINT(20)      NOT NULL COMMENT '更新时间(毫秒级时间戳)',
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Sepolia链NFT Item表';
# Item表通常是NFT项目中数据量最大、查询最频繁的表，以下索引对性能至关重要
-- 核心业务唯一约束：同一合约下的 token_id 必须唯一
ALTER TABLE `item_sepolia` ADD UNIQUE INDEX `uk_collection_token` (`collection_address`, `token_id`);

-- 按拥有者查询"我的NFT"（高频）
ALTER TABLE `item_sepolia` ADD INDEX `idx_owner` (`owner`);

-- 按所属 Collection 查询列表 + 分页排序
ALTER TABLE `item_sepolia` ADD INDEX `idx_collection_address` (`collection_address`);

-- 按上架时间倒序展示最新上架
ALTER TABLE `item_sepolia` ADD INDEX `idx_list_time` (`list_time`);

-- 按创建时间排序/分页
ALTER TABLE `item_sepolia` ADD INDEX `idx_create_time` (`create_time`);

###################################################################################
# 测试数据
INSERT INTO `item_sepolia` (
    `chain_id`, `collection_address`, `token_id`, `name`, `owner`, `creator`,
    `supply`, `list_price`, `list_time`, `sale_price`, `views`,
    `create_time`, `update_time`
) VALUES
-- 1. SEP-PUNKS #0001：已上架在售，有浏览量
(11155111, '0xA1B2C3D4E5F6A7B8C9D0E1F2A3B4C5D6E7F8A9B0', '1', 'Sepolia Punk #0001',
 '0xOwner1111111111111111111111111111111111', '0xCreator1111111111111111111111111111111',
 1, 0.0800000000, 1722932000000, 0.0000000000, 356,
 1722931200000, 1722932000000),

-- 2. SEP-PUNKS #0042：已售出，记录成交价
(11155111, '0xA1B2C3D4E5F6A7B8C9D0E1F2A3B4C5D6E7F8A9B0', '42', 'Sepolia Punk #0042',
 '0xOwner2222222222222222222222222222222222', '0xCreator1111111111111111111111111111111',
 1, 0.0000000000, 0, 0.1200000000, 1024,
 1722931500000, 1722940000000),

-- 3. NEW-APES #0001：新集合首枚，未上架，零浏览
(11155111, '0xB2C3D4E5F6A7B8C9D0E1F2A3B4C5D6E7F8A9B0C1', '1', 'New Sepolia Ape #0001',
 '0xCreator2222222222222222222222222222222', '0xCreator2222222222222222222222222222222',
 1, 0.0000000000, 0, 0.0000000000, 0,
 1722938400000, 1722938400000),

-- 4. SCAM-TOKEN #100：风险项目Item，低价上架
(11155111, '0xC3D4E5F6A7B8C9D0E1F2A3B4C5D6E7F8A9B0C1D2', '100', 'Fake Bored Ape #100',
 '0xOwner3333333333333333333333333333333333', '0xCreator3333333333333333333333333333333',
 1, 0.0010000000, 1722920000000, 0.0000000000, 88,
 1722844800000, 1722920000000),

-- 5. OLD-SEPOLIA #500：老项目高热度Item，多次转手
(11155111, '0xD4E5F6A7B8C9D0E1F2A3B4C5D6E7F8A9B0C1D2E3', '500', 'OG Sepolia Collectible #500',
 '0xOwner4444444444444444444444444444444444', '0xCreator4444444444444444444444444444444',
 1, 0.2000000000, 1722910000000, 0.1500000000, 5200,
 1720000000000, 1722910000000),

-- 6. GAME-ITEMS #1001：ERC1155多份发行，部分上架
(11155111, '0xE5F6A7B8C9D0E1F2A3B4C5D6E7F8A9B0C1D2E3F4', '1001', 'Iron Sword',
 '0xOwner5555555555555555555555555555555555', '0xCreator5555555555555555555555555555555',
 500, 0.0005000000, 1722935000000, 0.0000000000, 120,
 1722935000000, 1722935000000),

-- 7. GAME-ITEMS #2001：ERC1155未上架道具
(11155111, '0xE5F6A7B8C9D0E1F2A3B4C5D6E7F8A9B0C1D2E3F4', '2001', 'Golden Shield',
 '0xOwner6666666666666666666666666666666666', '0xCreator5555555555555555555555555555555',
 200, 0.0000000000, 0, 0.0000000000, 45,
 1722936000000, 1722936000000);