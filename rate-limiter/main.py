import json
import time
from redis_config import RedisConfig, RedisConnection, REDIS_HOST, REDIS_PORT, REDIS_DB
from algorithms.rate_limiter import (
    FixedWindowRateLimiter,
    SlidingWindowRateLimiterUsingList,
    SlidingWindowRateLimiterUsingSortedSet,
    SlidingWindowRateLimiterOptimized,
    SlidingWindowCounterRateLimiter,
    TokenBucketRateLimiter,
    LeakyBucketRateLimiter,
)

from enum import Enum

class RateLimiterAlgorithm(Enum):
    FIXED_WINDOW = "fixed_window"
    SLIDING_WINDOW_USING_LIST = "sliding_window_using_list"
    SLIDING_WINDOW_USING_SORTED_SET = "sliding_window_using_sorted_set"
    SLIDING_WINDOW_OPTIMIZED = "sliding_window_optimized"
    SLIDING_WINDOW_COUNTER = "sliding_window_counter"
    TOKEN_BUCKET = "token_bucket"
    LEAKY_BUCKET = "leaky_bucket"


if __name__ == "__main__":

    # Create Redis configuration
    redis_config = RedisConfig(REDIS_HOST, REDIS_PORT, REDIS_DB)
    print(redis_config)

    # Create Redis connection
    redis_connection = RedisConnection(redis_config)

    # # Get the connection
    # connection = redis_connection.get_connection()

    # # Use the connection (example: set and get a value)
    # connection.set("TEST_KEY", "123")
    # print("Value for 'TEST_KEY':", connection.get("TEST_KEY"))

    # object = {"name": "John Doe", "age": 30, "city": "New York"}

    # # Set an object in Redis
    # connection.set("TEST_OBJECT", json.dumps(object))
    # # Hash set an object in Redis
    # connection.hset("TEST_OBJECT_HASH", mapping=object)

    # # Disconnect
    # redis_connection.disconnect()

    # # Create Redis configuration
    # redis_config = RedisConfig(REDIS_HOST, REDIS_PORT, REDIS_DB)
    # print(redis_config)

    # # Create Redis connection
    # redis_connection = RedisConnection(redis_config)

    # # Get the connection
    # connection = redis_connection.get_connection()

    # prefix = "rate_limit"
    # algorithm = RateLimiterAlgorithm.FIXED_WINDOW.value
    # limit = 5
    # window_size = 10
    # user_id = "user_id_12345"

    # # Create a rate limiter
    # rate_limiter = FixedWindowRateLimiter(
    #     redis_connection, limit=limit, window_size=window_size
    # )

    # # Allow requests with a specific key
    # key = f"{prefix}:{algorithm}:{limit}:{window_size}:{user_id}"

    # # Test the rate limiter
    # for i in range(10):
    #     # time.sleep(1)
    #     if rate_limiter.allow_request(key):
    #         print(f"Request {i + 1} allowed")
    #     else:
    #         print(f"Request {i + 1} denied")

    # prefix = "rate_limit"
    # algorithm = RateLimiterAlgorithm.SLIDING_WINDOW_USING_LIST.value
    # limit = 10
    # window_size = 10
    # user_id = "user_id_12345"

    # key = f"{prefix}:{algorithm}:{limit}:{window_size}:{user_id}"

    # rate_limiter = SlidingWindowRateLimiterUsingList(
    #     redis_connection, limit=limit, window_size=window_size
    # )

    # # Test the rate limiter
    # for i in range(10):
    #     # time.sleep(1)
    #     if rate_limiter.allow_request(key):
    #         print(f"Request {i + 1} allowed")
    #     else:
    #         print(f"Request {i + 1} denied")

    # prefix = "rate_limit"
    # algorithm = RateLimiterAlgorithm.SLIDING_WINDOW_USING_SORTED_SET.value
    # limit = 10
    # window_size = 10
    # user_id = "user_id_12345"

    # key = f"{prefix}:{algorithm}:{limit}:{window_size}:{user_id}"

    # rate_limiter = SlidingWindowRateLimiterUsingSortedSet(
    #     redis_connection, limit=limit, window_size=window_size
    # )

    # # Test the rate limiter
    # for i in range(20):
    #     time.sleep(0.5)
    #     if rate_limiter.allow_request(key):
    #         print(f"Request {i + 1} allowed")
    #     else:
    #         print(f"Request {i + 1} denied")


    # prefix = "rate_limit"
    # algorithm = RateLimiterAlgorithm.SLIDING_WINDOW_OPTIMIZED.value
    # limit = 3000
    # window_size = 60
    # user_id = "user_id_12345"

    # key = f"{prefix}:{algorithm}:{limit}:{window_size}:{user_id}"

    # rate_limiter = SlidingWindowRateLimiterOptimized(
    #     redis_connection, limit=limit, window_size=window_size
    # )

    # # Test the rate limiter
    # for i in range(2000):
    #     time.sleep(0.001)
    #     if rate_limiter.allow_request(key):
    #         print(f"Request {i + 1} allowed")
    #     else:
    #         print(f"Request {i + 1} denied")


    # prefix = "rate_limit"
    # algorithm = RateLimiterAlgorithm.SLIDING_WINDOW_COUNTER.value
    # limit = 3
    # window_size = 5
    # user_id = "user_id_12345"

    # key = f"{prefix}:{algorithm}:{limit}:{window_size}:{user_id}"

    # rate_limiter = SlidingWindowCounterRateLimiter(
    #     redis_connection, limit=limit, window_size=window_size
    # )

    # # Test the rate limiter
    # for i in range(11):
    #     time.sleep(1)
    #     if rate_limiter.allow_request(key):
    #         print(f"TS: {i} sec | Request {i + 1} allowed")
    #     else:
    #         print(f"TS: {i} sec | Request {i + 1} denied")



    # prefix = "rate_limit"
    # algorithm = RateLimiterAlgorithm.TOKEN_BUCKET.value
    # limit = 10
    # refill_rate = 1
    # user_id = "user_id_12345"

    # key = f"{prefix}:{algorithm}:{limit}:{refill_rate}:{user_id}"

    # rate_limiter = TokenBucketRateLimiter(
    #     redis_connection, limit=limit, refill_rate=refill_rate
    # )

    # # Test the rate limiter
    # for i in range(20):
    #     time.sleep(0.1)
    #     if rate_limiter.allow_request(key):
    #         print(f"TS: {i} sec | Request {i + 1} allowed")
    #     else:
    #         print(f"TS: {i} sec | Request {i + 1} denied")
    #
    #


    prefix = "rate_limit"
    algorithm = RateLimiterAlgorithm.LEAKY_BUCKET.value
    limit = 10
    leak_rate = 1
    user_id = "user_id_12345"

    key = f"{prefix}:{algorithm}:{limit}:{leak_rate}:{user_id}"

    rate_limiter = LeakyBucketRateLimiter(
        redis_connection, limit=limit, leak_rate=leak_rate
    )

    # Test the rate limiter
    for i in range(20):
        time.sleep(0.1)
        if rate_limiter.allow_request(key):
            print(f"TS: {i} sec | Request {i + 1} allowed")
        else:
            print(f"TS: {i} sec | Request {i + 1} denied")
