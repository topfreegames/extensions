function create_cassandra_schema()
{
  echo "[INFO] create_cassandra_schema"
  local RESP=""
  while true; do
    RESP=$(docker exec -i extensions_cassandra_1 cqlsh -f ./create.cql)
    if [[ $RESP == "" ]]; then
      break
    fi

    echo "[INFO] $RESP"
    sleep 5
  done
}
