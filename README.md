# PgAuthProxy

[![Go Report Card](https://goreportcard.com/badge/github.com/KnifeMaster007/pgAuthProxy?style=flat-square)](https://goreportcard.com/report/github.com/KnifeMaster007/pgAuthProxy)

PgAuthProxy is a PostgreSQL gateway with customizable authentication. It provides 
single entrypoint for various database servers with credentials mapping capability.

## Installation

```
go install github.com/KnifeMaster007/pgAuthProxy
```

## Core concepts

When client connects PgAuthProxy, it handles the connection this way:

1. Receives startup message from the client, 
   which includes the names of the user and of the database
2. Responds to client with Authentication Request message
3. Client responds with password message
4. PgAuthProxy executes user-defined authenticator command
5. If authenticator command execution succeeded, 
   PgAuthProxy initiates connection to host, provided by authenticator
6. If connection to target succeeded, PgAuthProxy sends startup message 
   with parameters provided by the authenticator(user, database, etc.)
7. On authentication request from target host, PgAuthProxy sends 
   credential, provided by authenticator
8. If authentication with the target is successful, PgAuthProxy just forwards any 
   further messages between the client and target host

## Usage

```
pgAuthProxy [flags]

Flags:
      --clear-passwd    use cleartext password instead of MD5-hashed
      --config string   configuration file path
  -h, --help            help for pgAuthProxy
      --listen string   bind address (default ":5432")
```

## Configuration file

```yaml
listen: 0.0.0.0:15432                             # bind address (default ":5432")

authenticator:
  cleartext_password: false                       # use cleartext password instead of MD5-hashed
  cmd: ["authenticator.py", "--md5-passwords"]    # authentication command
```

## Authenticator executable call conventions

Authenticator must process startup message parameters and credentials, 
provided by user, and respond with startup message and credentials for target database server

PgAuthProxy launches authenticator on each connection, passes input parameters to STDIN
and reads target database parameters from an authenticator's STDOUT. 
Non-zero exit code treated as authentication error.

### Example

#### Authenticator input
```
user=testuser
database=testuser
application_name=psql
client_encoding=UTF8
_SOURCE_CRED=md55fa959c75491e1ce08541c50bc3ac3c4
_SOURCE_SALT=2182654f
```

If cleartext passwords is enabled, _SOURCE_CRED will contain password,
_SOURCE_SALT will be 00000000

#### Authenticator output
```
user=postgres
database=postgres
application_name=psql(proxied for testuser)
client_encoding=UTF8
_META_TARGET_HOST=pgbouncer.prod:5432
_META_TARGET_CRED=md53670464b1b43f39455d2637b187f9245
```

## Limitations
 * SSL is not supported yet
 * Cleartext password authentication with backend is not supported