CREATE TABLE `fileserver`.`fileserver`  (
  `id` varchar(36) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT 'ID',
  `file_md5` varchar(128) CHARACTER SET  COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '文件唯一md5值',
  `upload_name` varchar(36) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT '' COMMENT '上传至oss的文件名',
  `upload_number` int(11) NOT NULL DEFAULT 0 COMMENT '同一文件上传的个数',
  `upload_id` varchar(36) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT '' COMMENT '分块上传的uploadID',
  `oss_server` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT '' COMMENT '上传的oss服务',
  `create_at` bigint(20) NULL DEFAULT NULL COMMENT '创建时间',
  `update_at` bigint(20) NULL DEFAULT NULL COMMENT '修改时间',
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '文件上传表' ROW_FORMAT = DYNAMIC;