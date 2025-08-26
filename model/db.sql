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

CREATE TABLE `image` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `type` varchar(20) DEFAULT NULL COMMENT 'product 成品 poster 海报',
  `filename` varchar(100) DEFAULT NULL COMMENT '文件名',
  `url` varchar(255) DEFAULT NULL COMMENT '文件地址',
  `ip` varchar(20) DEFAULT NULL COMMENT '访问ip',
  `createTime` int(11) DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;