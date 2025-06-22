import time
from algorithms.base import RateLimiter
from redis_config import RedisConnection


class SlidingWindowCounterRateLimiter(RateLimiter):
    def __init__(self, redis_connection: RedisConnection, limit: int, window_size: int):
        self.redis_connection = redis_connection
        self.limit = limit
        self.window_size = window_size

    def allow_request(self, key: str) -> bool:
        """
        Determines whether a request is allowed based on the sliding window counter rate limiting logic.

        :param key: The unique key for the request (e.g., user ID, IP address).
        :return: True if the request is allowed, False otherwise.
        """
        lua_script = """
        local key = KEYS[1]
        local window_size = tonumber(ARGV[1])
        local limit = tonumber(ARGV[2])
        local current_time = tonumber(ARGV[3])

        -- Calculate the start of the window
        local current_window = math.floor(current_time / window_size)
        local previous_window = current_window - 1
        local elapsed_ratio = (current_time % window_size) / window_size

        -- Remove old windows that are older than prev window
        -- local old_windows = redis.call('HKEYS', key)
        -- for _, window in ipairs(old_windows) do
        --     if tonumber(window) < previous_window then
        --         redis.call('HDEL', key, window)
        --     end
        -- end


        -- Get the count of requests in the current and previous windows
        local current_count = tonumber(redis.call('HGET', key, current_window)) or 0
        local previous_count = tonumber(redis.call('HGET', key, previous_window)) or 0

        -- Calculate the effective count based on the elapsed ratio
        local effective_count = current_count + (previous_count * (1 - elapsed_ratio))

        -- If the effective count exceeds the limit, deny the request
        if effective_count >= limit then
            return false
        end

        -- Increment the count for the current window
        redis.call('HINCRBY', key, current_window, 1)

        -- Set the expiration for the key if it doesn't exist
        -- redis.call('EXPIRE', key, window_size * 10)
        return true
        """

        # Get the Redis connection
        connection = self.redis_connection.get_connection()

        # Get the current time
        current_time = int(time.time())

        # Execute the Lua script
        result = connection.eval(
            lua_script,
            1,  # Number of keys
            key,  # The key for the request
            self.window_size,  # Window size in seconds
            self.limit,  # Limit for the number of requests
            current_time  # Current time in seconds
        )

        # Return True if the request is allowed, False otherwise
        return bool(result)






"""
Simulate:
limit: 3 requests
window_size: 5 seconds

TS 1 sec:
previous window: 0
current window: 0
elapsed ratio: 0.2
effective_count = 0 + (0 * (1 - 0.2)) = 0
effective_count < limit, allow request
current_window: 1
-----------------------
TS 2 sec:
previous window: 0
current window: 1
elapsed ratio: 0.4
effective_count = 1 + (0 * (1 - 0.4)) = 1
effective_count < limit, allow request
current_window: 2
------------------------------
TS 3 sec:
previous window: 0
current window: 1
elapsed ratio: 0.6
effective_count = 2 + (0 * (1 - 0.6)) = 2
effective_count < limit, allow request
current_window: 3
------------------
TS 4 sec:
previous window: 0
current_window: 3
elapsed ratio: 0.8
effective_count = 3 + (0 * (1 - 0.8)) = 3
effective_count >= limit, deny request
-------------------
TS 5 sec:
previous window: 3
current_window: 0
elapsed ratio: 0.0
effective_count = 0 + (3 * (1 - 0.0)) = 3
effective_count >= limit, deny request
----------------
TS 6 sec:
previous window: 3
current_window: 0
elapsed ratio: 0.2
effective_count = 0 + (3 * (1 - 0.2)) = 2.4
effective_count < limit, allow request
current_window: 1
-----------------
TS 7 sec:
previous window: 3
current_window: 1
elapsed ratio: 0.4
effective_count = 1 + (3 * (1 - 0.4)) = 2.8
effective_count < limit, allow request
current_window: 2
------------------
TS 8 sec:
previous window: 3
current_window: 2
elapsed ratio: 0.6
effective_count = 2 + (3 * (1 - 0.6)) = 3.2
effective_count >= limit, deny request
------------------
TS 9 sec:
previous window: 3
current_window: 2
elapsed ratio: 0.8
effective_count = 3 + (3 * (1 - 0.8)) = 3.6
effective_count >= limit, deny request
------------------
TS 10 sec:
previous window: 2
current_window: 0
elapsed ratio: 0.0
effective_count = 0 + (2 * (1 - 0.0)) = 2
effective_count < limit, allow request






"""
