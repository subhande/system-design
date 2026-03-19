
# Remove docker container if it exists
docker rm -f rabbitmq

# Delete existing data directory if it exists
rm -rf rabbitmq-data

# Make data directory if it doesn't exist
mkdir -p rabbitmq-data

# Remove existing volume if it exists
docker volume rm -f rabbitmq_data


# Use Docker volume for data persistence in current directory/data
docker volume create \
    --name rabbitmq_data \
    --opt type=none \
    --opt device=$(pwd)/rabbitmq-data \
    --opt o=bind



docker run -d \
  --name rabbitmq \
  --env RABBITMQ_DEFAULT_USER=admin \
  --env RABBITMQ_DEFAULT_PASS=1234 \
  -p 5672:5672 \
  -p 15672:15672 \
  -v rabbitmq_data:/var/lib/rabbitmq \
  rabbitmq:4.2.5-management