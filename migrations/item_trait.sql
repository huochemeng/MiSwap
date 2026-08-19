CREATE TABLE `item_trait_sepolia`
(
    `id`                 BIGINT       NOT NULL AUTO_INCREMENT COMMENT '主键',
    `collection_address` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '所属Collection合约地址',
    `token_id`           VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'NFT Token ID',
    `trait`              VARCHAR(255) NOT NULL COMMENT '属性名称(如Background, Eyes)',
    `trait_value`        VARCHAR(255) NOT NULL COMMENT '属性值(如Blue, Laser)',
    `create_time`        BIGINT(20)   NOT NULL COMMENT '创建时间(毫秒级时间戳)',
    `update_time`        BIGINT(20)   NOT NULL COMMENT '更新时间(毫秒级时间戳)',
    PRIMARY KEY (`id`),
    -- 核心查询索引：按Collection+属性名+属性值统计稀有度/筛选
    INDEX `idx_collection_trait` (`collection_address`, `trait`, `trait_value`),
    -- 查询某个NFT的所有属性
    INDEX `idx_collection_token` (`collection_address`, `token_id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_unicode_ci COMMENT ='Sepolia链NFT属性特征表';

###################################################################################
# 测试数据
INSERT INTO `item_trait_sepolia` (
    `collection_address`, `token_id`, `trait`, `trait_value`,
    `create_time`, `update_time`
) VALUES
-- ========== SEP-PUNKS #1 (4个属性) ==========
('0xA1B2C3D4E5F6A7B8C9D0E1F2A3B4C5D6E7F8A9B0', '1', 'Background', 'Dark Purple', 1754478600000, 1754478600000),
('0xA1B2C3D4E5F6A7B8C9D0E1F2A3B4C5D6E7F8A9B0', '1', 'Skin', 'Zombie',      1754478600000, 1754478600000),
('0xA1B2C3D4E5F6A7B8C9D0E1F2A3B4C5D6E7F8A9B0', '1', 'Eyes',     'Laser',     1754478600000, 1754478600000),
('0xA1B2C3D4E5F6A7B8C9D0E1F2A3B4C5D6E7F8A9B0', '1', 'Mouth',    'Cigarette', 1754478600000, 1754478600000),

-- ========== SEP-PUNKS #42 (3个属性，含稀有属性 Alien Skin) ==========
('0xA1B2C3D4E5F6A7B8C9D0E1F2A3B4C5D6E7F8A9B0', '42', 'Background', 'Ocean Blue', 1754478700000, 1754478700000),
('0xA1B2C3D4E5F6A7B8C9D0E1F2A3B4C5D6E7F8A9B0', '42', 'Skin',     'Alien',      1754478700000, 1754478700000),
('0xA1B2C3D4E5F6A7B8C9D0E1F2A3B4C5D6E7F8A9B0', '42', 'Accessory','Gold Chain', 1754478700000, 1754478700000),

-- ========== NEW-APES #1 (5个属性) ==========
('0xB2C3D4E5F6A7B8C9D0E1F2A3B4C5D6E7F8A9B0C1', '1', 'Background', 'Sunset',    1754478800000, 1754478800000),
('0xB2C3D4E5F6A7B8C9D0E1F2A3B4C5D6E7F8A9B0C1', '1', 'Fur',        'Golden',    1754478800000, 1754478800000),
('0xB2C3D4E5F6A7B8C9D0E1F2A3B4C5D6E7F8A9B0C1', '1', 'Eyes',       'Bored',     1754478800000, 1754478800000),
('0xB2C3D4E5F6A7B8C9D0E1F2A3B4C5D6E7F8A9B0C1', '1', 'Mouth',      'Grin',      1754478800000, 1754478800000),
('0xB2C3D4E5F6A7B8C9D0E1F2A3B4C5D6E7F8A9B0C1', '1', 'Hat',        'Crown',     1754478800000, 1754478800000),

-- ========== OLD-SEPOLIA #500 (2个属性，极简风格) ==========
('0xD4E5F6A7B8C9D0E1F2A3B4C5D6E7F8A9B0C1D2E3', '500', 'Rarity',   'Legendary', 1754478600000, 1754478600000),
('0xD4E5F6A7B8C9D0E1F2A3B4C5D6E7F8A9B0C1D2E3', '500', 'Generation','OG',        1754478600000, 1754478600000),

-- ========== GAME-ITEMS #1001 (游戏道具属性) ==========
('0xE5F6A7B8C9D0E1F2A3B4C5D6E7F8A9B0C1D2E3F4', '1001', 'Type',     'Weapon',    1754479000000, 1754479000000),
('0xE5F6A7B8C9D0E1F2A3B4C5D6E7F8A9B0C1D2E3F4', '1001', 'Damage',   '45',        1754479000000, 1754479000000),
('0xE5F6A7B8C9D0E1F2A3B4C5D6E7F8A9B0C1D2E3F4', '1001', 'Durability','100',      1754479000000, 1754479000000),
('0xE5F6A7B8C9D0E1F2A3B4C5D6E7F8A9B0C1D2E3F4', '1001', 'Element',  'Fire',      1754479000000, 1754479000000);