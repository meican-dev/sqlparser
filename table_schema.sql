--this is a comment
DROP TABLE IF EXISTS `user`;
/* another comment */;
CREATE TABLE `user` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `username` varchar(20) DEFAULT NULL,
  `email` varchar(255) DEFAULT NULL,
  `birth_date` datetime NOT NULL DEFAULT '1970-01-01 00:00:00',
  `country_id` bigint(20) NOT NULL DEFAULT '0',
  `city_id` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `email` (`email`),
  KEY `FKBC63DCC747140EFE` (`city_id`),
  CONSTRAINT `FK5A735BAA2351BFBE` FOREIGN KEY (`city_id`) REFERENCES `city` (`id`),
  CONSTRAINT `FK5A735BAA2351BFBE` FOREIGN KEY (`country_id`) REFERENCES `country` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
