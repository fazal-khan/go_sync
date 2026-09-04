# go_sync

A data synchronization tool that reads from configured database tables, applies filter mutations to rows, and routes records to output targets.

## Language

**Table**:
A database table configured for synchronization, identified by a `database_name` and queried via a SQL statement.
_Avoid_: source, origin, endpoint

**Filter**:
Mutation rules applied to each row during ingestion — copy fields, remove fields, add fields, lowercase fields.
_Avoid_: transform, mutate rule, mapping

**Output**:
The destination a processed record is routed to: elasticsearch, api, file, or db.
_Avoid_: target, sink, destination

**Config**:
The XML configuration file defining all Tables, their Queries, Filters, and Outputs.
_Avoid_: settings, properties, configuration file

**Ingestion**:
The end-to-end process of fetching a Table's rows, applying its Filter, and routing the results to its Output.
_Avoid_: sync, pipeline, job
