CREATE KEYSPACE IF NOT EXISTS test
  WITH REPLICATION = {
    'class': 'SimpleStrategy',
    'replication_factor': 1
  };

CREATE TABLE IF NOT EXISTS test.user (
  id text,
  info text,
  PRIMARY KEY (id)
);
