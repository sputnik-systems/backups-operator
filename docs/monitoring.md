# Metrics
Appart from the general controller runtime metrics, operator exports following metrics:
* `backups_operator_backups` - each backup object corresponds to one metric. Metric supports these labels: `name` - object name, `namespace` - object namespace, `controller` - controller name (`clickhousebackup`, `dgraphbackup` for example), `status` - object status (`success` or `failed`).
* `backups_operator_scheduled_task_failures_total` - total count of failures in scheduled tasks execution. Metric labels: `name`, `namespace`, `controller`, `type` - schedule task type (`create` or `remove`).
