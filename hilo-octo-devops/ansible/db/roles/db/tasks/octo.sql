-- MySQL dump 10.13  Distrib 5.6.33, for Linux (x86_64)
--
-- Host: localhost    Database: octo
-- ------------------------------------------------------
-- Server version 5.6.33-log

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;

--
-- Table structure for table `apps`
--

DROP TABLE IF EXISTS `apps`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `apps` (
  `app_id` int(11) NOT NULL,
  `app_name` varchar(255) DEFAULT NULL,
  `description` varchar(255) DEFAULT NULL,
  `image_url` varchar(255) DEFAULT NULL,
  `email` varchar(255) DEFAULT NULL,
  `app_secret_key` varchar(255) CHARACTER SET utf8 COLLATE utf8_bin DEFAULT NULL,
  `client_secret_key` varchar(255) CHARACTER SET utf8 COLLATE utf8_bin DEFAULT NULL,
  `aes_key` char(32) DEFAULT NULL,
  `storage_type` int(11) DEFAULT NULL,
  PRIMARY KEY (`app_id`),
  UNIQUE KEY `app_secret_key` (`app_secret_key`),
  UNIQUE KEY `client_secret_key` (`client_secret_key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `file_urls`
--

DROP TABLE IF EXISTS `file_urls`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `file_urls` (
                             `app_id`       int(11) NOT NULL                                          DEFAULT '0',
                             `version_id`   int(11) NOT NULL                                          DEFAULT '0',
                             `revision_id`  int(11) NOT NULL                                          DEFAULT '0',
                             `object_name`  varchar(255) CHARACTER SET utf8 COLLATE utf8_bin NOT NULL DEFAULT '',
                             `crc`          bigint(20)                                                DEFAULT NULL,
                             `md5`          varchar(255)                                              DEFAULT NULL,
                             `url`          varchar(1024)                                             DEFAULT NULL,
                             `state`        int(11)                                                   DEFAULT NULL,
                             `upd_datetime` datetime                                                  DEFAULT NULL,
                             PRIMARY KEY (`app_id`,`version_id`,`object_name`,`revision_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `files`
--

DROP TABLE IF EXISTS `files`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `files` (
  `id` int(11) NOT NULL DEFAULT '0',
  `app_id` int(11) NOT NULL DEFAULT '0',
  `version_id` int(11) NOT NULL DEFAULT '0',
  `revision_id` int(11) DEFAULT NULL,
  `filename` varchar(255) DEFAULT NULL,
  `object_name` varchar(16) CHARACTER SET utf8 COLLATE utf8_bin DEFAULT NULL,
  `url` varchar(1024) DEFAULT NULL,
  `size` int(11) DEFAULT NULL,
  `crc` bigint(20) DEFAULT NULL,
  `generation` bigint(20) DEFAULT NULL,
  `md5` varchar(255) DEFAULT NULL,
  `tag` varchar(255) DEFAULT NULL,
  `dependency` varchar(6553) DEFAULT NULL,
  `priority` int(11) DEFAULT NULL,
  `state` int(11) DEFAULT NULL,
  `build_number` varchar(20) DEFAULT NULL,
  `upload_version_id` int(11) DEFAULT NULL,
  `upd_datetime` datetime DEFAULT NULL,
  PRIMARY KEY (`app_id`,`version_id`,`id`),
  UNIQUE KEY `object_name` (`app_id`,`version_id`,`object_name`),
  UNIQUE KEY `filename` (`app_id`,`version_id`,`filename`),
  KEY `idx_files_appid_versionid_revisionid` (`app_id`,`version_id`,`revision_id`),
  KEY `object_name_2` (`object_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `gcs`
--

DROP TABLE IF EXISTS `gcs`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `gcs` (
  `app_id` int(11) NOT NULL DEFAULT '0',
  `project_id` varchar(255) NOT NULL DEFAULT '',
  `backet` varchar(255) DEFAULT NULL,
  `location` varchar(16) DEFAULT NULL,
  PRIMARY KEY (`app_id`,`project_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `resource_urls`
--

DROP TABLE IF EXISTS `resource_urls`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `resource_urls` (
                                 `app_id`       int(11) NOT NULL                                          DEFAULT '0',
                                 `version_id`   int(11) NOT NULL                                          DEFAULT '0',
                                 `revision_id`  int(11) NOT NULL                                          DEFAULT '0',
                                 `object_name`  varchar(255) CHARACTER SET utf8 COLLATE utf8_bin NOT NULL DEFAULT '',
                                 `md5`          varchar(255)                                              DEFAULT NULL,
                                 `url`          varchar(1024)                                             DEFAULT NULL,
                                 `state`        int(11)                                                   DEFAULT NULL,
                                 `upd_datetime` datetime                                                  DEFAULT NULL,
                                 PRIMARY KEY (`app_id`,`version_id`,`object_name`,`revision_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `resources`
--

DROP TABLE IF EXISTS `resources`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `resources` (
  `id` int(11) NOT NULL DEFAULT '0',
  `app_id` int(11) NOT NULL DEFAULT '0',
  `version_id` int(11) NOT NULL DEFAULT '0',
  `revision_id` int(11) DEFAULT NULL,
  `filename` varchar(255) DEFAULT NULL,
  `object_name` varchar(16) CHARACTER SET utf8 COLLATE utf8_bin DEFAULT NULL,
  `url` varchar(1024) DEFAULT NULL,
  `size` int(11) DEFAULT NULL,
  `generation` bigint(20) DEFAULT NULL,
  `md5` varchar(255) DEFAULT NULL,
  `tag` varchar(255) DEFAULT NULL,
  `priority` int(11) DEFAULT NULL,
  `state` int(11) DEFAULT NULL,
  `build_number` varchar(20) DEFAULT NULL,
  `upload_version_id` int(11) DEFAULT NULL,
  `upd_datetime` datetime DEFAULT NULL,
  PRIMARY KEY (`app_id`,`version_id`,`id`),
  UNIQUE KEY `object_name` (`app_id`,`version_id`,`object_name`),
  UNIQUE KEY `filename` (`app_id`,`version_id`,`filename`),
  KEY `object_name_2` (`object_name`),
  KEY `idx_resources_appid_versionid_revisionid` (`app_id`,`version_id`,`revision_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `secret`
--

DROP TABLE IF EXISTS `secret`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `secret` (
  `app_id` int(11) NOT NULL,
  `secret` varchar(255) NOT NULL DEFAULT '',
  `auth_type` int(11) DEFAULT '0',
  `upd_datetime` datetime DEFAULT NULL,
  PRIMARY KEY (`secret`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `tags`
--

DROP TABLE IF EXISTS `tags`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `tags` (
  `app_id` int(11) NOT NULL DEFAULT '0',
  `tag_id` int(11) NOT NULL DEFAULT '0',
  `name` varchar(255) CHARACTER SET utf8 COLLATE utf8_bin DEFAULT NULL,
  PRIMARY KEY (`app_id`,`tag_id`),
  UNIQUE KEY `app_id` (`app_id`,`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `user_apps`
--

DROP TABLE IF EXISTS `user_apps`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `user_apps` (
  `app_id` int(11) NOT NULL,
  `user_id` varchar(255) NOT NULL,
  `role_type` int(11) DEFAULT '0',
  PRIMARY KEY (`app_id`,`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `users`
--

DROP TABLE IF EXISTS `users`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `users` (
  `user_id` varchar(255) NOT NULL,
  `password` varchar(255) DEFAULT NULL,
  `email` varchar(255) DEFAULT NULL,
  `auth_type` int(11) DEFAULT '0',
  PRIMARY KEY (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `versions`
--

DROP TABLE IF EXISTS `versions`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `versions` (
                            `app_id`          int(11) NOT NULL DEFAULT '0',
                            `version_id`      int(11) NOT NULL DEFAULT '0',
                            `description`     varchar(255)     DEFAULT NULL,
                            `max_revision`    int(11)          DEFAULT NULL,
                            `copy_version_id` int(11)          DEFAULT NULL,
                            `copy_app_id`     int(11)          DEFAULT NULL,
                            `env_id`          int(11)          DEFAULT NULL,
                            `state`           int(11)          DEFAULT NULL,
                            `upd_datetime`    datetime         DEFAULT NULL,
                            PRIMARY KEY (`app_id`,`version_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

DROP TABLE IF EXISTS `envs`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `envs` (
  `app_id` int(11) NOT NULL,
  `env_id` int(11) AUTO_INCREMENT NOT NULL,
  `name` varchar(200) NOT NULL,
  `detail` varchar(2000) DEFAULT NULL,
  PRIMARY KEY (`env_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

--
-- Table structure for table `buckets`
--

DROP TABLE IF EXISTS `buckets`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `buckets` (
  `app_id` varchar(255) NOT NULL,
  `bucket_name` varchar(255) NOT NULL,
  PRIMARY KEY (`app_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

/*!40101 SET character_set_client = @saved_cs_client */;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;


-- Dump completed on 2016-09-15  2:08:16
