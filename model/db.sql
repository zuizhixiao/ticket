CREATE TABLE `template` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `userId` int(11) NOT NULL DEFAULT '0' COMMENT '用户id 为0代表系统',
  `url` varchar(250) NOT NULL DEFAULT '' COMMENT '模板图片',
  `titleColor` varchar(20) NOT NULL DEFAULT '' COMMENT '电影标题颜色',
  `textColor` varchar(20) NOT NULL DEFAULT '' COMMENT '电影详情颜色',
  `titleFontSize` int(4) NOT NULL COMMENT '电影标题字体大小',
  `textFontSize` int(4) NOT NULL COMMENT '电影详情字体大小',
  `nameFontSize` int(4) NOT NULL COMMENT '专属名称字体大小',
  `tailFontSize` int(4) NOT NULL COMMENT '末尾文字字体大小',
  `status` tinyint(1) NOT NULL DEFAULT '1' COMMENT '1正常 2删除',
  `createTime` int(11) NOT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_user` (`userId`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=21 DEFAULT CHARSET=utf8 COMMENT='模板库';