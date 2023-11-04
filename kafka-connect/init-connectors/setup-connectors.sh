# Launch Kafka Connect
/etc/confluent/docker/run &
#
# Wait for Kafka Connect listener
echo "Waiting for Kafka Connect to start listening on localhost ⏳"
while : ; do
  curl_status=$$(curl -s -o /dev/null -w %{http_code} http://localhost:8083/connectors)
  echo -e $$(date) " Kafka Connect listener HTTP state: " $$curl_status " (waiting for 200)"
  if [ $$curl_status -eq 200 ] ; then
    break
  fi
  sleep 5 
done

echo -e "\n--\n+> Creating MQTT Source Connect"
curl -s -X PUT -H  "Content-Type:application/json" http://localhost:8083/connectors/mqtt-source-connect/config \
    -d '@mqtt-source-connect.json'

echo -e "\n--\n+> Creating Debezium Postgres Source Connect"
curl -s -X PUT -H  "Content-Type:application/json" http://localhost:8083/connectors/debezium-connector-postgresql/config \
    -d '@debezium-connector-postgresql.json'
sleep infinity