

from db import create_pool
import random
ranges_pool = None
ids = list(range(10))


async def get_next_id():
    """
    Generate a unique ID
    """
    global ranges_pool
    if ranges_pool is None:
        ranges_pool = await create_pool(db_name="ranges")

    id = random.choice(ids)
    result = None
    async with ranges_pool.acquire() as conn:
        async with conn.cursor() as cursor:
            # Start a transaction
            await conn.begin()
            query = f"""SELECT * from ranges WHERE id = {id}"""
            await cursor.execute(query)
            result = await cursor.fetchone()

            query = f"""UPDATE ranges SET current = current + 1 WHERE id = {id}"""
            await cursor.execute(query)

            # Commit the transaction
            await conn.commit()

    if result is None:
        raise Exception("No result found")

    return result[-1]
