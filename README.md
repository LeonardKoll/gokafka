# Kafka & Go

This repository contains the result of introducing myself to Apache Kafka & Go.

## What it does & Technology

The producer module contains a simple file listener. On change, it will send the contents of that file to Kafka.


The sqladapter module connect to kafka and expects messages in the form indicated in example.json. The adapter persists the information in a Postgres-DB (hosted on RDS).


The nosqladapter is similar to the sqladapter but persists information in a DynamoDB instead.

## Kafka

Set up Kafka and create a topic 'testTopic':
```
bin/kafka-topics.sh --create --zookeeper localhost:2181 --replication-factor 1 --partitions 1 --topic testTopic
```

## producer configuration

Place a config.json file in the producer folder and populate accordingly:
```
{
    "AwsTable" : "Your DynamoDB Table Name",
    "Topic" : "testTopic",
    "Partition" : 0,
    "KafkaHost" : "hostname:port",
    "Offset" : 0
}
```

## sqladapter configuration

Place a config.json file in the sqladapter folder and populate accordingly:
```
{
    "DbConnection" : "host=YOURHOST.rds.amazonaws.com port=5432 user=postgres dbname=YOURDB password=YOURPASSWORD",
    "Topic" : "testTopic",
    "Partition" : 0,
    "KafkaHost" : "hostname:port",
    "Offset" : 0
}
```

## nosqladapter configuration

Place a config.json file in the nosqladapter folder and populate accordingly:
```
{
    "Topic" : "testTopic",
    "Partition" : 0,
    "KafkaHost" : "hostname:port",
    "WatchFile" : "C:\\some\\folders\\watch.json"
}
```

## AWS credentials

Make sure you have configured the shared credentials and configuration files on you local machine.
For more information:
https://docs.aws.amazon.com/ses/latest/DeveloperGuide/create-shared-credentials-file.html
https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-files.html


## Postgres DB

You may set up the DB using the scripts below. However, Gorm (Go library for ORM) will do that automatically for you during the frist write.

```
-- Table: public.addresses

-- DROP TABLE public.addresses;

CREATE TABLE public.addresses
(
    id integer NOT NULL DEFAULT nextval('addresses_id_seq'::regclass),
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    street text COLLATE pg_catalog."default",
    zip text COLLATE pg_catalog."default",
    city text COLLATE pg_catalog."default",
    user_id integer,
    CONSTRAINT addresses_pkey PRIMARY KEY (id)
)

TABLESPACE pg_default;

ALTER TABLE public.addresses
    OWNER to postgres;
-- Index: idx_addresses_deleted_at

-- DROP INDEX public.idx_addresses_deleted_at;

CREATE INDEX idx_addresses_deleted_at
    ON public.addresses USING btree
    (deleted_at ASC NULLS LAST)
    TABLESPACE pg_default;
```

```
-- Table: public.users

-- DROP TABLE public.users;

CREATE TABLE public.users
(
    id integer NOT NULL DEFAULT nextval('users_id_seq'::regclass),
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    name text COLLATE pg_catalog."default",
    phone text COLLATE pg_catalog."default",
    CONSTRAINT users_pkey PRIMARY KEY (id)
)

TABLESPACE pg_default;

ALTER TABLE public.users
    OWNER to postgres;
-- Index: idx_users_deleted_at

-- DROP INDEX public.idx_users_deleted_at;

CREATE INDEX idx_users_deleted_at
    ON public.users USING btree
    (deleted_at ASC NULLS LAST)
    TABLESPACE pg_default;
```