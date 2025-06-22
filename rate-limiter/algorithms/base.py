from abc import ABC, abstractmethod
from redis_config import RedisConnection


class RateLimiter(ABC):
    def __init__(self, redis_connection: RedisConnection, limit: int, window_size: int):
        self.redis_connection = redis_connection
        self.limit = limit
        self.window_size = window_size

    @abstractmethod
    def allow_request(self, key: str) -> bool:
        """
        Determines whether a request is allowed based on the rate limiting logic.

        :param key: The unique key for the request (e.g., user ID, IP address).
        :return: True if the request is allowed, False otherwise.
        """
        pass
