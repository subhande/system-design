import mysql.connector

# Establish the connection
connection = mysql.connector.connect(
    host='localhost',        # e.g., 'localhost'
    user='root',    # e.g., 'root'
    password='12345678',
    database='demo' # e.g., 'test_db'
)

# Create a cursor object
cursor = connection.cursor()

# Example to check connection
cursor.execute("SELECT DATABASE();")
db_name = cursor.fetchone()
print("Connected to database:", db_name)
cursor.execute("SHOW DATABASES;")
databases = cursor.fetchall()
print("Databases:", databases)


######################################################################
####################   range TABLE   ################################
######################################################################

query = "CREATE DATABASE IF NOT EXISTS ranges"
cursor.execute(query)
print("Database created successfully")

query = "USE ranges"
cursor.execute(query)
print("Database selected successfully")


query = "DROP TABLE IF EXISTS ranges"
cursor.execute(query)
query = "CREATE TABLE IF NOT EXISTS ranges (id INT PRIMARY KEY, start INT, end INT, current INT)"
cursor.execute(query)
print("Table created successfully")

offset = 100000
current = 0
for i in range(10):
    query = f"INSERT INTO ranges (id, start, end, current) VALUES ({i}, {current}, {current+offset}, {current})"
    current += offset
    cursor.execute(query)

query = "SELECT * FROM ranges"
cursor.execute(query)
result = cursor.fetchall()
for row in result:
    print(row)


######################################################################
####################   URL TABLE   ##################################
######################################################################

query = "CREATE DATABASE IF NOT EXISTS urls"
cursor.execute(query)
print("Database created successfully")

query = "USE urls"
cursor.execute(query)
print("Database selected successfully")

query = "DROP TABLE IF EXISTS urls"
cursor.execute(query)
query = "CREATE TABLE IF NOT EXISTS urls (short_code VARCHAR(10) PRIMARY KEY, long_url TEXT, expiration_date DATETIME)"
cursor.execute(query)
print("Table created successfully")

# Close cursor and connection
cursor.close()
connection.close()
