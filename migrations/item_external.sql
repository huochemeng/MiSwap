CREATE TABLE `item_external_sepolia`
(
    `id`                  BIGINT       NOT NULL AUTO_INCREMENT COMMENT '主键',
    `collection_address`  VARCHAR(255) NOT NULL DEFAULT '' COMMENT '所属Collection合约地址',
    `token_id`            VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'NFT Token ID',
    `is_uploaded_oss`     TINYINT(1)   NOT NULL DEFAULT 0 COMMENT '图片是否已上传OSS(0:未上传,1:已上传)',
    `upload_status`       INT          NOT NULL DEFAULT 0 COMMENT '图片上传状态(0:待处理,1:成功,2:失败,3:重试中)',
    `meta_data_uri`       VARCHAR(1024) NOT NULL DEFAULT '' COMMENT '元数据URI(IPFS/HTTP)',
    `image_uri`           VARCHAR(1024) NOT NULL DEFAULT '' COMMENT '原始图片URI',
    `oss_uri`             VARCHAR(1024) NOT NULL DEFAULT '' COMMENT 'OSS图片CDN地址',
    `is_video_uploaded`   TINYINT(1)   NOT NULL DEFAULT 0 COMMENT '视频是否已上传OSS(0:未上传,1:已上传)',
    `video_upload_status` INT          NOT NULL DEFAULT 0 COMMENT '视频上传状态(0:待处理,1:成功,2:失败,3:重试中)',
    `video_type`          VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '视频MIME类型(video/mp4等)',
    `video_uri`           VARCHAR(1024) NOT NULL DEFAULT '' COMMENT '原始视频URI',
    `video_oss_uri`       VARCHAR(1024) NOT NULL DEFAULT '' COMMENT 'OSS视频CDN地址',
    `create_time`         BIGINT(20)   NOT NULL COMMENT '创建时间(毫秒级时间戳)',
    `update_time`         BIGINT(20)   NOT NULL COMMENT '更新时间(毫秒级时间戳)',
    PRIMARY KEY (`id`),
    UNIQUE INDEX `uk_collection_token` (`collection_address`, `token_id`),
    INDEX `idx_upload_status` (`upload_status`),
    INDEX `idx_video_upload_status` (`video_upload_status`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_unicode_ci COMMENT ='NFT外部资源OSS上传状态表';

###################################################################################
# 测试数据
INSERT INTO `item_external_sepolia` (
    `collection_address`, `token_id`,
    `is_uploaded_oss`, `upload_status`, `meta_data_uri`, `image_uri`, `oss_uri`,
    `is_video_uploaded`, `video_upload_status`, `video_type`, `video_uri`, `video_oss_uri`,
    `create_time`, `update_time`
) VALUES
-- 1. SEP-PUNKS #1：图片+视频均已上传成功
('0xA1B2C3D4E5F6A7B8C9D0E1F2A3B4C5D6E7F8A9B0', '1',
 1, 1, 'ipfs://QmPunk001/metadata.json', 'ipfs://QmPunk001/image.png', 'https://cdn.example.com/sepolia/punks/1.png',
 1, 1, 'video/mp4', 'ipfs://QmPunk001/video.mp4', 'https://cdn.example.com/sepolia/punks/1.mp4',
 1754478600000, 1754478600000),

-- 2. SEP-PUNKS #42：仅图片上传成功，无视频
('0xA1B2C3D4E5F6A7B8C9D0E1F2A3B4C5D6E7F8A9B0', '42',
 1, 1, 'ipfs://QmPunk042/metadata.json', 'ipfs://QmPunk042/image.png', 'https://cdn.example.com/sepolia/punks/42.png',
 0, 0, '', '', '',
 1754478700000, 1754478700000),

-- 3. NEW-APES #1：图片上传失败，需要重试
('0xB2C3D4E5F6A7B8C9D0E1F2A3B4C5D6E7F8A9B0C1', '1',
 0, 2, 'ipfs://QmApe001/metadata.json', 'ipfs://QmApe001/image.png', '',
 0, 0, '', '', '',
 1754478800000, 1754478900000),

-- 4. OLD-SEPOLIA #500：图片成功，视频上传失败
('0xD4E5F6A7B8C9D0E1F2A3B4C5D6E7F8A9B0C1D2E3', '500',
 1, 1, 'ipfs://QmOG500/metadata.json', 'ipfs://QmOG500/image.gif', 'https://cdn.example.com/sepolia/og/500.gif',
 0, 2, 'video/webm', 'ipfs://QmOG500/video.webm', '',
 1754478600000, 1754479000000),

-- 5. GAME-ITEMS #1001：纯图片资源，全部成功
('0xE5F6A7B8C9D0E1F2A3B4C5D6E7F8A9B0C1D2E3F4', '1001',
 1, 1, 'https://api.game.com/meta/sword.json', 'https://api.game.com/img/sword.png', 'https://cdn.example.com/game/sword.png',
 0, 0, '', '', '',
 1754479000000, 1754479000000),

-- 6. GAME-ITEMS #2001：待处理状态 ⏳（刚入库尚未触发上传）
('0xE5F6A7B8C9D0E1F2A3B4C5D6E7F8A9B0C1D2E3F4', '2001',
 0, 0, 'https://api.game.com/meta/shield.json', 'https://api.game.com/img/shield.png', '',
 0, 0, '', '', '',
 1754479100000, 1754479100000),

-- 7. SCAM-TOKEN #100：图片重试中
('0xC3D4E5F6A7B8C9D0E1F2A3B4C5D6E7F8A9B0C1D2', '100',
 0, 3, 'ipfs://QmFake100/metadata.json', 'ipfs://QmFake100/image.jpg', '',
 0, 0, '', '', '',
 1754479200000, 1754479300000);