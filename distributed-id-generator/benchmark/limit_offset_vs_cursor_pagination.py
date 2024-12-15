
import mysql.connector
import time

# 1️⃣ Connect to the MySQL server
connection = mysql.connector.connect(
    host="localhost",
    user="root",
    password="root",
    database="counter_service_db"
)

# 2️⃣ Create a cursor object
cursor = connection.cursor()

NO_OF_ROWS = 6000000

# Limit Pagination Query
query_pagination = "SELECT * FROM user2 LIMIT <LIMIT> OFFSET <OFFSET>"

# Using Cursor Based Query
query_cursor = "SELECT * FROM user2 WHERE id > <LAST_ID> LIMIT <LIMIT>"

LIMIT = 100

BATCH_SIZE = 1000000

limit_offset_pagination_timings = []
cursor_based_pagination_timings = []
offsets = []

for i in range(0, NO_OF_ROWS, BATCH_SIZE):
    offset = i
    print(f"Offset: {offset}")
    offsets.append(offset)

    start = time.time()
    
    # 3️⃣ Execute the query
    query = query_pagination.replace("<LIMIT>", str(BATCH_SIZE)).replace("<OFFSET>", str(offset))
    cursor.execute(query)

    # 4️⃣ Fetch the result
    result = cursor.fetchall()
    # print(result[0])
    
    end = time.time()
    
    duration = round((end-start) * 1000)
    
    limit_offset_pagination_timings.append(duration)
    
    print(f"LIMIT OFFSET Pagination: Time Taken: {duration} ms for {LIMIT} records with offset {offset}")
    
    
    start = time.time()
    
    query = query_cursor.replace("<LIMIT>", str(LIMIT)).replace("<LAST_ID>", str(offset))
    
    cursor.execute(query)
    
    result = cursor.fetchall()
    # print(result[0])
    
    end = time.time()
    
    duration = round((end-start) * 1000)
    
    print(f"Cursor Based Pagination: Time Taken: {duration} ms for {LIMIT} records with last id {offset}")
    
    cursor_based_pagination_timings.append(duration)
    



    """
        # Limit Offset Pagination
        When we use LIMIT and OFFSET in our query, MySQL will first fetch all the records and then apply the LIMIT and OFFSET clause.
        This is not efficient as MySQL will fetch all the records and then apply the LIMIT and OFFSET clause.
        Performance will degrade as the offset increases.
        
        # Cursor Based Pagination
        In Cursor Based Pagination, we use the last id from the previous batch to fetch the next batch of records.
        This is more efficient as MySQL will fetch only the required records.
        Index is required. This will only work if it is Full Index Query and we need sequential data from the Index.

    """
    
