# Backup types
Right now operator supports backuping:
* `Dgraph` - through dgraph [export creation request](https://dgraph.io/docs/deploy/dgraph-administration/#export-database). Implemented only s3 storage.
* `ClickHouse` - through [clickhouse-backup](https://github.com/AlexAkulov/clickhouse-backup).

# Dgraph Backup
You can create dgraph backup by creating `DgraphBackup` object. Example:
```
apiVersion: backups.sputnik.systems/v1alpha1
kind: DgraphBackup
metadata:
  name: dgraph-1634289801
spec:
  adminUrl: http://dgraph-dgraph-alpha:8080/admin
  destination: s3://s3.us-east-2.amazonaws.com/dgraph-test
  region: us-east-2
  secrets:
    - dgraph-backup-s3-creds
```
* `adminUrl` - is url of dgraph cluster admin. If object is in the same namespace, you can skip namespace specification in admin url.
* `destination` - bucket url
* `region` - required for cleanup tasks successfully execution.

# Dgraph Backup Schedule
`DgraphBackupSchedule` may be used for periodically create and rotate backup objects. Example:
```
apiVersion: backups.sputnik.systems/v1alpha1
kind: DgraphBackupSchedule
metadata:
  name: default
  spec:
    schedule: "*/10 * * * *"
    retention: 30m
    backup:
      adminUrl: http://dgraph-dgraph-alpha:8080/admin
      destination: s3://s3.us-east-2.amazonaws.com/dgraph-test
      region: us-east-2
      secrets:
        - dgraph-backup-s3-creds
```
* `backup` - same as `DgraphBackup` object `spec` field.
* `schedule` - backup creation schedule in cron notation(supports `@every`, `@weekly`, `@daily` etc).
* `retention` - lifetime of backup objects managed by this schedule object.

# ClickHouse Backup
`ClickHouseBackup` object creates ClickHouse backup:
```
apiVersion: backups.sputnik.systems/v1alpha1
kind: ClickHouseBackup
metadata:
  name: clickhousebackup-sample
  spec:
    apiAddress: http://chi-default-default-0-0:7171
```
* `apiAddress` - clickhouse-backup api address. Namespace postfix can be omitted, if api runned in same namespace.
* `createParams` - create request params kv.
* `uploadParams` - upload request params kv.

# ClickHouse Backup Schedule
`ClickHouseBackupSchedule` fields equal `DgraphBackupSchedule` object fileds, `spec.backup` will be copy-pasted into `ClickHouseBackup` `spec` field:
```
apiVersion: backups.sputnik.systems/v1alpha1
kind: ClickHouseBackupSchedule
metadata:
  name: clickhousebackupschedule-sample
spec:
  schedule: "*/5 * * * *"
  retention: 15m
  backup:
    apiAddress: http://chi-default-default-0-0:7171
```
