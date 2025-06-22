import time
from algorithms.base import RateLimiter
from redis_config import RedisConnection

class LeakyBucketRateLimiter(RateLimiter):
    def __init__(self, redis_connection: RedisConnection, limit: int, leak_rate: int):
        self.redis_connection = redis_connection
        self.limit = limit
        self.leak_rate = leak_rate

    def allow_request(self, key: str) -> bool:
        """
        Determines whether a request is allowed based on the leaky bucket rate limiting logic.

        :param key: The unique key for the request (e.g., user ID, IP address).
        :return: True if the request is allowed, False otherwise.
        """

        LEAKY_BUCKET_LUA_SCRIPT = """

        -- Fetch current state from Redis hash
        local key = KEYS[1]
        local now = tonumber(ARGV[1])
        local capacity = tonumber(ARGV[2])
        local leak_rate = tonumber(ARGV[3])

        local state = redis.call('HMGET', key, 'current_level', 'last_leak_timestamp')
        local current_level = tonumber(state[1]) or 0
        local last_leak_ts = tonumber(state[2]) or now

        -- Calculate how much has leaked since last check
        local leaked = (now - last_leak_ts) * leak_rate
        local new_level = math.max(0, current_level - leaked)

        local allowed = false
        if new_level < capacity then
            new_level = new_level + 1
            allowed = true
        end

        -- Save new state
        redis.call('HMSET', key, 'current_level', new_level, 'last_leak_timestamp', now)

        -- Set TTL (optional, e.g. 2x idle time)
        redis.call('EXPIRE', key, 3600)

        return allowed
        """

        # Get the Redis connection
        connection = self.redis_connection.get_connection()

        # Get the current time
        current_time = int(time.time())


        # Execute the Lua script
        result = connection.eval(
            LEAKY_BUCKET_LUA_SCRIPT,
            1,  # Number of keys
            key,  # The key for the leaky bucket
            current_time,  # Current time
            self.limit,  # Bucket capacity
            self.leak_rate  # Leak rate
        )


        # Return the result of the Lua script
        return bool(result)
