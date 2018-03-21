# watchdog
This program is a watchdog for monitoring and restarting linux services. It is integrated with AWS DynamoDB from which it loads configuration. It also sends notifications about services state through AWS SnS.
## usage
watchdog supports the following flags:  
--logs-dir - path to directory containing log files  
--logfile-split-threshold - when logfile size will be bigger than this value, then old content of log will be moved to separate file and sent to s3 bucket  
--s3-bucket - name of s3 bucket for storing logs  
--dynamo-table - name of table in DynamoDB storing configuration for watchdog  
--dynamo-key - value of primary key of a record in DynamoDB table storing configuration for watchdog  
--sns - name of SnS topic used to receive and forward notifications emitted by watchdog  
