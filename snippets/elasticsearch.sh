# Remove existing container if it exists
docker rm -f snippets-elasticsearch

# Delete existing data directory if it exists
rm -rf es-data

# Make data directory if it doesn't exist
mkdir -p es-data

# Fix permissions (important for Elasticsearch)
chmod -R 777 es-data

# Remove existing volume if it exists
docker volume rm -f snippets-elasticsearch-data

# Create Docker volume mapped to local directory
docker volume create \
    --name snippets-elasticsearch-data \
    --opt type=none \
    --opt device=$(pwd)/es-data \
    --opt o=bind

# Run Elasticsearch with security enabled + custom password
docker run --name snippets-elasticsearch \
    -p 9200:9200 \
    -p 9300:9300 \
    -e discovery.type=single-node \
    -e ELASTIC_PASSWORD=StrongPassword123 \
    -e xpack.security.enabled=true \
    -e ES_JAVA_OPTS="-Xms512m -Xmx512m" \
    -v snippets-elasticsearch-data:/usr/share/elasticsearch/data \
    -d elasticsearch:8.19.4


# Remove existing container if it exists
docker rm -f snippets-kibana

# Wait for Elasticsearch to be ready
until curl -s -u elastic:StrongPassword123 http://localhost:9200 > /dev/null 2>&1; do
    sleep 2
done

# Generate service account token
TOKEN=$(docker exec snippets-elasticsearch bin/elasticsearch-service-tokens create elastic/kibana default | awk -F'= ' '{print $2}')

# Run Kibana with service account token
docker run --name snippets-kibana \
    --link snippets-elasticsearch:elasticsearch \
    -p 5601:5601 \
    -e ELASTICSEARCH_HOSTS=http://host.docker.internal:9200 \
    -e ELASTICSEARCH_SERVICEACCOUNTTOKEN=$TOKEN \
    -d docker.elastic.co/kibana/kibana:8.19.4