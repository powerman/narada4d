# Narada4D [![GoDoc](https://godoc.org/github.com/powerman/narada4d/schemaver?status.svg)](https://godoc.org/github.com/powerman/narada4d/schemaver) [![Go Report Card](https://goreportcard.com/badge/github.com/powerman/narada4d)](https://goreportcard.com/report/github.com/powerman/narada4d) [![CircleCI](https://circleci.com/gh/powerman/narada4d.svg?style=svg)](https://circleci.com/gh/powerman/narada4d) [![Coverage Status](https://coveralls.io/repos/github/powerman/narada4d/badge.svg?branch=master)](https://coveralls.io/github/powerman/narada4d?branch=master)

Applications like stateful microservices deployed in docker containers
needs to manage their data schema version - both to protect against
occasional damaging data because of running container with older
application version on data volume or database with newer data schema
version, and to provide a reliable way to migrate data schema between
versions. Sometimes to reliably and effectively backup your data you also
need to know it schema version.

Narada4D defines a way to manage your data schema version and safely
access/migrate/backup/restore your data.

It's based on open protocols (algorithms) which describe where you can
store your data schema version (for ex. in a file or SQL database) and how
it can be reliably locked, get and set.

Also it provides some basic tools and libraries to make it easier to
manage data schema version, with support for some of these protocols.

It was designed to be very flexible and extensible, so feel free to write
your own implementations (tools and libraries) for same protocols or
design new protocols - as long as they follow same base rules they all
should be compatible and keep your data safe.

# Workflow

- Application must use versioning for it data schema.
- To access own data application must:
    - Acquire lock (usually shared, but some operations may require
      exclusive) on data schema version.
    - Read current data schema version.
        - It is guaranteed data schema version won't change until lock
          will be released.
    - Access data in case of supported data schema version.
    - Release lock.
- In case application see unsupported data schema version it may either
  exit or just repeat until this change.

# Requirements

- Application's data schema has own versioning.
    - *Rationale:* It's too easy to occasionally and sometimes makes sense
      to intentionally run container with older app using volume with data
      from newer app. This is usual case and must be supported, thus app
      must know data schema version(s) it's able to handle and check it.
- Application shouldn't keep shared lock on data schema version for a
  long enough time. Usually application should either acquire and release
  shared lock for each data access. If this is too ineffective then it
  should acquire it on start and then release and acquire again every few
  seconds (to give a chance for other tool to acquire exclusive lock).
    - *Rationale:* We need to get exclusive lock to make consistent backup
      (and migrate data schema) while application is running.
- Environment variable `$NARADA4D` must contains URL to data schema
  version. Schema in this URL (like `file:` or `mysql:`) selects Narada4D
  protocol which should be used for managing data schema version.
  Tools/libraries should return error for unsupported protocols.
    - *Rationale:* This is convenient and flexible way to configure
      application, libraries and tools in one place.
- Environment variable `$NARADA4D_SKIP_LOCK` must be set to non-empty
  string when exclusive lock was acquired, and if it's set then next
  attempt to acquire and then release shared or exclusive lock shouldn't
  do anything. It's recommended to set it to space-separated list of
  URLs to data schema version - in this case exclusive lock won't do
  anything only if one of these URLs will match URL we're locking.
    - *Rationale:* Needed to support recursive locking, needed in case
      when one tool (migrate) runs other tools (backup or restore).
- Version of data schema must be a string which is either `none` or
  `dirty` or consists of one or more digits separated with single dots
  (for ex. `42`, `0`, `0.12.0`).
    - *Rationale:* This make it easier to use it as part of backup file
      name and store in other places.
- Version of data schema must be set to `dirty` in case operation which
  modify data schema (migration, restoring backup) was interrupted and
  current data may not conform to any known data schema versions.
    - *Rationale:* This prevent application from using damaged data
      schema until it will be fixed (manually, or by restoring backup).
- Version of data schema must be set to `none` after initializing version
  value, before defining first data schema.
    - *Rationale:* This enables to separate initialization of data schema
      version from other operations.

## Recommendations

- One application have only one data set and thus one data schema version.
    - *Rationale:* This will ensure there will be no deadlocks while
      trying to lock multiple data sets.
- It usually doesn't makes any sense to keep application version and data
  schema version in sync, or even use same versioning style for both.
    - *Rationale:* Data schema changes less often than application, so
      it's usually convenient to version it using just single number.

# Protocols

Managing data schema versions requires:

- Ability to acquire shared and exclusive lock on it.
- Automatically releasing lock in case process acquired this lock has
  exited or crashed.
- Guarantee exclusive lock will be acquired ASAP, even in case new shared
  locks always requested before releasing all current shared locks.

## file:///path/to/dir

- Based on flock(2).
    - *Rationale:* More than one application needs simultaneous access to
      data (at least - main application and backup tool). Some of them may
      be running in another container or (when data is on network FS) at
      another server.
    - In case of using NFS file locks must be global (not local to current
      host): don't use mount options `nolock` or `local_lock` with any
      values except `none`.
- All path names mentioned below are relative to path in `$NARADA4D`.
- `.version`
    - Symlink to current data schema version.
        - *Rationale:* Symlink can be read/written using one atomic
          syscall.
    - While it's not exists no one (except initialization tool) should do
      anything with files in this directory, including attempts to acquire
      locks.
    - Created after successful initialization.
    - Never removed.
    - Modified only under exclusive lock on `.lock`.
        - *Rationale:* This guarantee data schema version won't change
          while application hold shared or exclusive lock on `.lock`.
- `.lock` and `.lock.queue`
    - Usual, empty files.
    - Created while initialization.
    - Never removed.
        - *Rationale:* This makes possible to open these files just once
          when application start and then lock/unlock already open files
          (this speedup locking in about 4 time).
    - Any data access (excluding test for `.version` existence but
      including reading `.version`) is allowed only under shared or
      exclusive lock on `.lock`.
    - Before trying to acquire shared or exclusive lock on `.lock` it's
      required to acquire exclusive lock on `.lock.queue` first, which
      should be released immediately after acquiring lock on `.lock`.
        - *Rationale:* It guarantee exclusive lock on `.lock` will be
          acquired ASAP.
- Typical initialization flow:
    - Create directory from `$NARADA4D` if it's not exists.
    - Ensure `.version` is not exists or exit.
    - Create empty usual file `.lock` or ensure it's already exists.
    - Create empty usual file `.lock.queue` or ensure it's already exists.
    - Create `.version` symlink to `none`.
- Typical application flow:
    - On start:
        - Ensure `.version` is exists or exit.
        - Open `.lock` and `.lock.queue` to speedup locking.
    - On data access:
        - Acquire exclusive lock on `.lock.queue`.
        - Acquire shared or exclusive lock on `.lock`.
            - Exclusive lock is required for operations which may change
              data schema version (like migration or restoring backup) and
              also may be required to make consistent backup.
        - Release lock on `.lock.queue`.
        - Read `.version`.
        - If version supported then access data, else either exit or
          release lock in next step and try again later.
        - Release lock on `.lock`.
    - As each data access require 5 extra syscalls applications with high
      data access rate (about 30000 RPS) may like to acquire lock on start
      and then release and immediately re-acquire it every second,
      delaying data access meanwhile.

## mysql://user[:pass]@host[:port]/database

- Version is stored in table named `Narada4D`, in a row `var="version"`.
- Neither table nor this row is never deleted.
- To initialize: `CREATE TABLE Narada4D (var VARCHAR(191) PRIMARY KEY, val
  VARCHAR(255) NOT NULL) SELECT "version" as var, "none" as val`.
- To check is it initialized: `SELECT COUNT(*) FROM Narada4D`.
- To set shared lock: `LOCK TABLE Narada4D READ`.
- To set exclusive lock: `LOCK TABLE Narada4D WRITE`.
- To unlock: `UNLOCK TABLES`.
- To get version: `SELECT val FROM Narada4D WHERE var='version'`.
- To change version: `UPDATE Narada4D SET val=? WHERE var='version'`.

## goose-mysql://user[:pass]@host[:port]/database

The `goose` tool is not aware about Narada4D and it'll manage DB schema
version on it's own. Because of this `SchemaVer.Set` is not supported with
`goose`.

- Version is stored in table named `goose_db_version`.
- This table is managed by [goose](https://github.com/pressly/goose) tool.
- Second table named `Narada4D` is used for locking and is never deleted.
- To initialize: call any goose command/API plus `CREATE TABLE Narada4D
  (var VARCHAR(191) PRIMARY KEY, val VARCHAR(255) NOT NULL) SELECT
  "version_from" as var, "goose" as val`.
- To check is it initialized: `SELECT COUNT(*) FROM Narada4D`.
- To set shared lock: `LOCK TABLE Narada4D READ`.
    - This won't prevent *anyone* not aware about Narada4D (including
      `goose` tool) from making changes.
- To set exclusive lock: `LOCK TABLE Narada4D WRITE`.
    - This won't prevent *anyone* not aware about Narada4D (including
      `goose` tool) from making changes, but only one of Narada4D-aware
      apps will be running after acquiring this lock.
- To unlock: `UNLOCK TABLES`.
- To get version: call goose API.
- To change version: call goose command to apply some up/down migration.
- **TODO:** It is unclear how to manage "dirty" in case goose fail some
  migration which was executed not within transaction (goose doesn't
  provide any way to detect this case and continue to report previous
  schema version after failed migration, which is incorrect).
    - It's recommended to keep [statements that cause an implicit commit](https://dev.mysql.com/doc/refman/5.7/en/implicit-commit.html)
      (like CREATE/ALTER/DROP/TRUNCATE TABLE) in their own migrations,
      with **one statement per migration**.

## goose-postgres://user[:pass]@host[:port]/database?sslmode=disable&…

The `goose` tool is not aware about Narada4D and it'll manage DB schema
version on it's own. Because of this `SchemaVer.Set` is not supported with
`goose`.

- Version is stored in table named `goose_db_version`.
- This table is managed by [goose](https://github.com/pressly/goose) tool.
- To initialize: call any goose command/API.
- To check is it initialized: `SELECT COUNT(*) FROM goose_db_version`.
- To set shared lock: `LOCK TABLE goose_db_version IN SHARE MODE`.
    - This will prevent `goose` tool from making any changes (actually
      it'll hang waiting for lock, so make sure you have set corresponding
      timeouts, e.g. PostgreSQL `statement_timeout`) but allows it to read
      current version/status.
- To set exclusive lock: `LOCK TABLE goose_db_version IN SHARE UPDATE EXCLUSIVE MODE`.
    - This won't prevent *anyone* not aware about Narada4D (including
      `goose` tool) from making changes, but only one of Narada4D-aware
      apps will be running after acquiring this lock.
- To unlock: commit/rollback transaction used to set lock.
    - Make sure transaction won't be closed prematurely because of idle
      timeout.
- To get version: call goose API.
- To change version: call goose command to apply some up/down migration.
- **TODO:** It is unclear how to manage "dirty" in case goose fail some
  migration which was executed not within transaction (goose doesn't
  provide any way to detect this case and continue to report previous
  schema version after failed migration, which is incorrect).

# Tools

## narada4d-init

Initialize schema version at location provided in $NARADA4D.
Schema version must not be already initialized.

## narada4d-lock [/path/to/cmd [args…]]

Run given command with given args under exclusive lock on schema version
at location provided in $NARADA4D. Without command will run shell (useful
for manual maintenance in case of "dirty" schema version).
Exits with exit code of executed command or 127 of command was terminated
by signal.
