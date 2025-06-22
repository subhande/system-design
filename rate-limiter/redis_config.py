"""
brew --version
brew install redis
redis-server
brew services start redis
brew services info redis
brew services stop redis
redis-cli
"""

import redis

REDIS_HOST = "localhost"
REDIS_PORT = 6379
REDIS_DB = 0


class RedisConfig:
    def __init__(self, host: str, port: int, db: int):
        self.host = host
        self.port = port
        self.db = db

    def __repr__(self):
        return f"RedisConfig(host={self.host}, port={self.port}, db={self.db})"


class RedisConnection:
    def __init__(self, config: RedisConfig):
        self.config = config
        self.connection = None
        self.connect()

    def connect(self):
        # Connect To Redis
        self.connection = redis.StrictRedis(
            host=self.config.host,
            port=self.config.port,
            db=self.config.db,
            decode_responses=True,
        )
        # Test the connection
        try:
            self.connection.ping()
            print("Connected to Redis")
        except redis.ConnectionError as e:
            print(f"Failed to connect to Redis: {e}")
            self.connection = None

    def get_connection(self):
        if self.connection is None:
            raise Exception("Connection not established. Call connect() first.")
        return self.connection

    def disconnect(self):
        if self.connection:
            self.connection.close()
            self.connection = None
        else:
            print("No connection to close.")
