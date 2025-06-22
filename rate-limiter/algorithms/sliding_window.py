import time
from algorithms.base import RateLimiter
from redis_config import RedisConnection
import uuid

class SlidingWindowRateLimiterUsingList(RateLimiter):
    def __init__(self, redis_connection: RedisConnection, limit: int, window_size: int):
        self.redis_connection = redis_connection
        self.limit = limit
        self.window_size = window_size

    def allow_request(self, key: str) -> bool:
        """
        Determines whether a request is allowed based on the sliding window rate limiting logic.

        :param key: The unique key for the request (e.g., user ID, IP address).
        :return: True if the request is allowed, False otherwise.
        """

        lua_script = """
        local key = KEYS[1]
        local window_size = tonumber(ARGV[1])
        local limit = tonumber(ARGV[2])
        local current_time = tonumber(ARGV[3])

        local start_time = current_time - window_size

        -- Get list of timestamps from Redis --
        local timestamps = redis.call('LRANGE', key, 0, -1)

        -- Remove timestamps older than the window size from LIST --
        local valid_timestamps = {}
        for i, timestamp in ipairs(timestamps) do
            local ts = tonumber(timestamp)
            if ts and ts >= start_time then
                table.insert(valid_timestamps, ts)
            end
        end

        local count = #valid_timestamps
        -- If the count exceeds the limit, deny the request --
        if count >= limit then
            return false
        end


        -- Delete the old timestamps from the list --
        redis.call('DEL', key)
        -- Add back the valid timestamps and the current timestamp --
        if count > 0 then
            redis.call('LPUSH', key, unpack(valid_timestamps))
        end
        redis.call('LPUSH', key, current_time)
        -- Set the expiration time for the key --
        redis.call('EXPIRE', key, window_size)
        return true
        """

        # Get the Redis connection
        connection = self.redis_connection.get_connection()

        # Get the current time in seconds
        current_time = int(time.time())

        # Execute the Lua script
        result = connection.eval(
            lua_script, 1, key, self.window_size, self.limit, current_time
        )

        return bool(result)


"""

Sliding Window Rate Limiting using Sorted Set in Redis
Window Size: 10 seconds
Limit: 3 requests per window

"1003" (score: 1003)
"1006" (score: 1006)
"1012" (score: 1012)


| Time (s) | Request | Action Taken                             | Members After    |
| -------- | ------- | ---------------------------------------- | ---------------- |
| 1000     | 1       | ZADD (1000, "1000")                      | 1000             |
| 1003     | 2       | ZADD (1003, "1003")                      | 1000, 1003       |
| 1006     | 3       | ZADD (1006, "1006")                      | 1000, 1003, 1006 |
| 1009     | 4       | ZCARD == 3 ⇒ REJECT                      | 1000, 1003, 1006 |
| 1012     | 5       | ZREMRANGEBYSCORE ≤ 1002, ZADD (1012,...) | 1003, 1006, 1012 |

ZADD (1000, "1000-uuid")
- uuid is a unique identifier for the request, which can be used to track individual requests.
- Without uuid, the score is just the timestamp, which can lead to issues if multiple requests are made at the same second it will be overwritten and counted as one request.


"""


class SlidingWindowRateLimiterUsingSortedSet(RateLimiter):
    def __init__(self, redis_connection: RedisConnection, limit: int, window_size: int):
        self.redis_connection = redis_connection
        self.limit = limit
        self.window_size = window_size

    def allow_request(self, key: str) -> bool:
        """
        Determines whether a request is allowed based on the sliding window rate limiting logic.

        :param key: The unique key for the request (e.g., user ID, IP address).
        :return: True if the request is allowed, False otherwise.
        """

        lua_script = """
        local key = KEYS[1]
        local window_size = tonumber(ARGV[1])
        local limit = tonumber(ARGV[2])
        local current_time = tonumber(ARGV[3])
        local req_id = ARGV[4]

        local start_time = current_time - window_size

        -- Remove timestamps older than the window size from the sorted set --
        redis.call('ZREMRANGEBYSCORE', key, 0, start_time)

        -- Get the count of timestamps in the sorted set --
        local count = redis.call('ZCARD', key)

        -- If the count exceeds the limit, deny the request --
        if count >= limit then
            return false
        end

        -- Add the current timestamp to the sorted set with the current time as the score --
        redis.call('ZADD', key, current_time, req_id)
        redis.call('EXPIRE', key, window_size)
        return true
        """

        # Get the Redis connection
        connection = self.redis_connection.get_connection()

        # Get the current time in seconds
        current_time = int(time.time())

        # Generate a unique request ID
        req_id = f"{current_time}-{uuid.uuid4().hex}"

        # Execute the Lua script
        result = connection.eval(
            lua_script, 1, key, self.window_size, self.limit, current_time, req_id
        )

        return bool(result)



class SlidingWindowRateLimiterOptimized(RateLimiter):
    def __init__(self, redis_connection: RedisConnection, limit: int, window_size: int, granularity: int = 1):
        self.redis_connection = redis_connection
        self.limit = limit
        self.window_size = window_size
        self.granularity = granularity

    def allow_request(self, key: str) -> bool:
        """
        Determines whether a request is allowed based on the sliding window rate limiting logic.

        :param key: The unique key for the request (e.g., user ID, IP address).
        :return: True if the request is allowed, False otherwise.
        """

        lua_script = """
        local key = KEYS[1]
        local window_size = tonumber(ARGV[1])
        local limit = tonumber(ARGV[2])
        local current_time = tonumber(ARGV[3])

        local start_time = current_time - window_size

        -- Get mapping for the key --
        local mapping = redis.call('HGETALL', key)

        -- {timestamp: count} mapping --
        -- Remove older timestamps from the mapping --
        local updated_mapping = {}
        local total_requests = 0
        for i = 1, #mapping, 2 do
            local timestamp = mapping[i]
            local count = mapping[i+1]
            local ts = tonumber(timestamp)
            if ts and ts >= start_time then
                table.insert(updated_mapping, ts)
                table.insert(updated_mapping, count)
                total_requests = total_requests + tonumber(count)
            end
        end

        -- If the total requests exceed the limit, deny the request --
        if total_requests >= limit then
            return false
        end

        -- Update the mapping then increment the count for the current timestamp --
        if #updated_mapping > 0 then
            redis.call('HSET', key, unpack(updated_mapping))
        end

        -- Add the current timestamp with count 1 --
        redis.call('HINCRBY', key, current_time, 1)  -- Increment the count for the current timestamp
        -- Set the expiration time for the key --
        redis.call('EXPIRE', key, window_size)
        return true
        """

        # Get the Redis connection
        connection = self.redis_connection.get_connection()

        # Get the current time in seconds
        current_time = int(time.time())

        # Adjust the current time based on the granularity
        # This ensures that requests are counted in fixed intervals
        current_time = current_time - (current_time % self.granularity)

        # Execute the Lua script
        result = connection.eval(
            lua_script, 1, key, self.window_size, self.limit, current_time
        )

        return bool(result)
