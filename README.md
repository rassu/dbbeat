# Database Beat

Elastic Beats to enable Postgres CRUD events (only ships events about insert, update and delete operations and no ships of read operations) to sync Elastic index with
Postgres database. Tool will create a trigger to trace database
transactions and send it to Elasticsearch index that has name
convention `{database name}-{table name}`
(e.g. for a table named `test` in `db` database the index name will be `db-test`).

## Table of Content
1. [Database Beat](#database-beat)
1. [Developer Guide](#developer-guide)


## Developer Guide
[beats developer guide](https://www.elastic.co/guide/en/beats/libbeat/current/new-beat.html).


### Setup
- [go 1.15 or higher](https://golang.org/doc/install)
- mage *(optional)*
  * Run command: `make install-mage`

- docker and docker-compose *(for test purpose only)*
  * We used docker to run Postgres, Elasticsearch and Kibana (you can just run command `make docker`):

### Build Beats

On root of the repository, run `mage build` or `make build`, which should output a binary named `dbbeat`. `./dbbeat -e -d "*"` for verbose logs in stdout, or simply run `./dbbeat` if you prefer not to see any log.

### Relevant Directories and Files
#### ./dbbeat.yml
For Beat configuration.

- Database connection can be changed from `uri` attribute e.g.:
```yaml
dbbeat:
  db_config:
    uri: "postgres://postgres:pwd@localhost:5432/db?sslmode=disable"
```
where the connection format is as follows:
`postgres://[user[:password]@][netloc][:port][/dbname][?param1=value1&...]`

- You can change Elasticsearch connection in the `hosts` attribute e.g.:
```yaml
output.elasticsearch:
  # Array of hosts to connect to.
  hosts: ["localhost:9200"]
```
#### ./beater
The main workhorse, connects to Postgres and ships event data to Elasticsearch
#### ./postgres/
For Postgres setup. initially wanted this to be `./db` directory, to allow connecting to various databases, but due to time and scope constraint, limits to Postgres only
- Creates trigger for selected tables and operations (defaults to all tables and CRUD operations)
- Creates notify_trigger that will send notification to channel on relevant table's operation
- Creates listeners for selected tables and operations (defaults to all tables and CRUD operations)
#### ./examples [FIXME]
example usage of populating Postgres database and doing operations, not industry standard, use your own migration tools

### Testing Steps

1. Insert/Update/Delete a record from the database.
    - create tables: `make test-create-db`
    - insert data: `make test-insert-db`
1. Assuming you run Elasticsearch and Kibana following the above setup,
open `http://localhost:5601`

2. Enter username/password: elastic/changeme
and you should see something like:
```yellow open db-names UmcF96y1QA-TOU6cQGO-dw 1 1 1 0 20.1kb 20.1kb```
where `db-names` is the name of the index where the beats are shipped into.
   Note: index name consists of `{database name}-{table name}`.

1. To see changes log you can check the index, e.g. the following is a curl command should return channel, ops, and payloads of listened operations.
```shell
curl -X GET "http://localhost:9200/db-names/_search?pretty" -H 'Content-Type: application/json' -d' { "_source": { "includes": [ "payload.*", "channel", "ops" ] } } '
```

