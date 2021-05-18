# SLOW QUERY
# Time: 180208  5:19:02
# User@Host: octo[octo] @  [10.102.14.9]  Id:  9353
# Query_time: 0.195625  Lock_time: 0.000053 Rows_sent: 38155  Rows_examined: 38155
#SET timestamp=1518067142;
#SELECT app_id,version_id,id,revision_id,filename,object_name,size,crc,generation,md5,tag,dependency,priority,state,upd_datetime
#FROM files where app_id=31 and version_id=12 order by revision_id DESC;
ALTER TABLE `octo`.`files`
ADD INDEX `revision_id_desc` (`app_id` ASC, `version_id` ASC, `revision_id` DESC);