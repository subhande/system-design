import time
from algorithms.base import RateLimiter
from redis_config import RedisConnection


class TokenBucketRateLimiter(RateLimiter):
    def __init__(self, redis_connection: RedisConnection, limit: int, refill_rate: int):
        self.redis_connection = redis_connection
        self.limit = limit
        self.refill_rate = refill_rate

    def allow_request(self, key: str) -> bool:
        """
        Determines whether a request is allowed based on the token bucket rate limiting logic.

        :param key: The unique key for the request (e.g., user ID, IP address).
        :return: True if the request is allowed, False otherwise.
        """

        TOKEN_BUCKET_LUA_SCRIPT = """
            local key = KEYS[1]
            local capacity = tonumber(ARGV[1])
            local refill_rate = tonumber(ARGV[2])
            local now = tonumber(ARGV[3])
            local requested = tonumber(ARGV[4])

            -- Get the bucket
            local data = redis.call("HMGET", key, "tokens", "timestamp")
            local tokens = tonumber(data[1])
            local timestamp = tonumber(data[2])

            -- If the bucket does not exist, initialize it
            if tokens == nil then
                tokens = capacity
                timestamp = now
            end

            -- Add tokens since last check
            local delta = math.max(0, now - timestamp)
            local refill = delta * refill_rate
            tokens = math.min(capacity, tokens + refill)
            timestamp = now

            local allowed = false
            if tokens >= requested then
                allowed = true
                tokens = tokens - requested
            end

            -- Store the new state
            redis.call("HMSET", key, "tokens", tokens, "timestamp", timestamp)
            -- Set TTL (optional, e.g. 2x idle time)
            redis.call("EXPIRE", key, 3600)
            return allowed

        """

        # Get the Redis connection
        connection = self.redis_connection.get_connection()

        # Get the current time
        current_time = int(time.time())

        # Execute the Lua script
        result = connection.eval(
            TOKEN_BUCKET_LUA_SCRIPT, 1, key, self.limit, self.refill_rate, current_time, 1
        )

        return bool(result)
