# watchdog
This program is a watchdog for monitoring and restarting linux services. It is integrated with AWS DynamoDB from which it loads configuration. It also sends notifications about services state through AWS SnS.
## usage
watchdog supports the following flags:  
--logfile - path to log file. If a logfile can't be read or created, then stdout will be used to write messages  
--dynamo-table - name of table in DynamoDB storing configuration for watchdog  
--dynamo-key - value of primary key of a record in DynamoDB table storing configuration for watchdog  
--sns - name of SnS topic used to receive and forward notifications emited by watchdog  
