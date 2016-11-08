#!/bin/sh
export AWS_ACCESS_KEY_ID=xxxxxxxxxxxxxx
export AWS_SECRET_ACCESS_KEY=xxxxxxxxxxxxxx
export AWS_REGION=ap-northeast-1

# clear backups
rm /backups/data/*

# archive
cd /
tar zcvfp /backups/data/etc.tar.gz /etc
tar zcvfp /backups/data/home.tar.gz /home
tar zcvfp /backups/data/var.tar.gz /var

# backup to S3
/backups/backup-to-s3 /backups/data bucket host