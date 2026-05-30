# Remove existing container if it exists
docker rm -f lumo-drive-postgres

# Delete existing data directory if it exists
rm -rf postgres-data

# Make data directory if it doesn't exist
mkdir -p postgres-data

# Remove existing volume if it exists
docker volume rm -f lumo-drive-postgres-data

# Use Docker volume for data persistence in current directory/data
docker volume create \
    --name lumo-drive-postgres-data \
    --opt type=none \
    --opt device=$(pwd)/postgres-data \
    --opt o=bind
docker run --name lumo-drive-postgres \
    -e POSTGRES_PASSWORD=postgres \
    -e POSTGRES_USER=postgres \
    -e POSTGRES_DB=lumo_drive_db \
    -p 5432:5432 \
    -v lumo-drive-postgres-data:/var/lib/postgresql \
    -d postgres:18.3
