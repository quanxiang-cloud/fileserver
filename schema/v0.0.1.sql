CREATE TABLE `user`  (
  `id` varchar(36) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL COMMENT 'ID',
  `file_name` varchar(300) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL DEFAULT '' COMMENT '文件名称',
  `file_md5` varchar(128) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL DEFAULT '' COMMENT '文件唯一md5值',
  `file_type` varchar(200) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL COMMENT '文件mime类型',
  `file_time` int(11) NOT NULL DEFAULT 0 COMMENT '视频文件长度 单位秒',
  `file_size` int(11) NOT NULL DEFAULT 0 COMMENT '文件大小 单位KB',
  `file_number` int(11) NOT NULL DEFAULT 1 COMMENT '同一文件上传的个数',
  `upload_name` varchar(300) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL DEFAULT '' COMMENT '文件上传保存的文件名',
  `url` varchar(500) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL DEFAULT '' COMMENT 'URL地址',
  `create_at` bigint(20) NOT NULL COMMENT '创建时间',
  `update_at` bigint(20) NOT NULL COMMENT '修改时间',
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8 COLLATE = utf8_general_ci COMMENT = '文件上传表' ROW_FORMAT = DYNAMIC;
