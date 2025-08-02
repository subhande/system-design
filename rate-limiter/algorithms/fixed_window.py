import time
from algorithms.base import RateLimiter
from redis_config import RedisConnection


class FixedWindowRateLimiter(RateLimiter):
    def __init__(self, redis_connection: RedisConnection, limit: int, window_size: int):
        self.redis_connection = redis_connection
        self.limit = limit
        self.window_size = window_size

    def allow_request_naive(self, key: str) -> bool:
        """
        Determines whether a request is allowed based on a naive fixed window rate limiting logic.

        :param key: The unique key sfor the request (e.g., user ID, IP address).
        :return: True if the request is allowed, False otherwise.

        This naive implementation uses Redis' INCR command to count requests and checks against the limit.
        Without Lua scripting, it may lead to inconsistant behavior in a distributed environment
        """

        # # This part is optional, but it helps to ensure that the key is unique for the current time window
        # current_window = int(time.time() // self.window_size) * self.window_size
        # # Create a unique key for the current window
        # key = f"{key}:{current_window}"

        # Get the Redis connection
        connection = self.redis_connection.get_connection()

        # Increment the count for the key
        current_count = connection.incr(key)
        current_count = int(current_count)  # Ensure the count is an integer
        print(f"Current count for key '{key}': {current_count}")

        # If the count exceeds the limit, deny the request
        if current_count > self.limit:
            return False

        # If this is the first request, set the expiration time for the key
        if current_count == 1:
            connection.expire(key, self.window_size)
        return True

    def allow_request(self, key: str) -> bool:
        """
        Determines whether a request is allowed based on the fixed window rate limiting logic.

        :param key: The unique key for the request (e.g., user ID, IP address).
        :return: True if the request is allowed, False otherwise.
        """

        # # This part is optional, but it helps to ensure that the key is unique for the current time window
        # current_window = int(time.time() // self.window_size) * self.window_size
        # # Create a unique key for the current window
        # key = f"{key}:{current_window}"

        lua_script = """
        local key = KEYS[1]
        local window_size = tonumber(ARGV[1])
        local limit = tonumber(ARGV[2])
        local current_count = redis.call('INCR', key)
        if current_count == 1 then
            redis.call('EXPIRE', key, window_size)
        end
        return current_count <= limit
        """
        # Get the Redis connection
        connection = self.redis_connection.get_connection()

        # Execute the Lua script
        result = connection.eval(lua_script, 1, key, str(self.window_size), str(self.limit))
        return bool(result)
