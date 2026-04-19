import aiomysql

async def create_pool(db_name: str):
    # Creating an async connection pool
    pool = await aiomysql.create_pool(
        host='localhost',  # Database host
        port=3306,         # Port (default MySQL port)
        user='root',  # MySQL username
        password='12345678',  # MySQL password
        db=db_name,  # Database name
        minsize=1,         # Minimum number of connections in the pool
        maxsize=10,        # Maximum number of connections in the pool
    )
    return pool
