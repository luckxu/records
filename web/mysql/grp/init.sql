set sql_log_bin=0;
create user 'grp'@'%';
alter user 'grp'@'%' identified by 'grp12345678';
grant REPLICATION SLAVE on *.* to 'grp'@'%';
grant BACKUP_ADMIN on *.* to 'grp'@'%';
flush privileges;
set sql_log_bin=1;
change master to master_user='grp', master_password='grp12345678' for channel 'group_replication_recovery';
install plugin group_replication soname 'group_replication.so';
