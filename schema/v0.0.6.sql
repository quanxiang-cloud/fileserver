-- 把表默认的字符集和所有字符列（CHAR,VARCHAR,TEXT）改为新的字符集
ALTER TABLE `fileserver`.`fileserver` CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 修改
ALTER TABLE `fileserver`.`fileserver` change `file_md5` `digest` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '文件唯一md5值';
ALTER TABLE `fileserver`.`fileserver` change `file_type` `path` varchar(300) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '存储服务中的路径';
ALTER TABLE `fileserver`.`fileserver` change `file_number` `number` int NOT NULL DEFAULT '1' COMMENT '同一文件上传的个数';
ALTER TABLE `fileserver`.`fileserver` change `upload_name` `upload_id` varchar(300) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '分块上传uploadID';
ALTER TABLE `fileserver`.`fileserver` change `file_size` `store_name` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '使用的存储配置名';
ALTER TABLE `fileserver`.`fileserver` change `file_time` `upload_type` int NOT NULL COMMENT '上传的类型，1：文件上传,2：自定义页面上传';

-- drop
ALTER TABLE `fileserver`.`fileserver` DROP `id`;
ALTER TABLE `fileserver`.`fileserver` DROP `url`;
-- 查看相关表与字段编码
-- SHOW CREATE TABLE `user`;
-- SHOW FULL COLUMNS FROM `user`;