import pika, json, time
from faker import Faker

# Initialize Faker to generate fake data
fake = Faker()

# Set the RabbitMQ server credentials
credentials = pika.PlainCredentials("admin", "1234")

# Establish a connection to the RabbitMQ server
connection = pika.BlockingConnection(pika.ConnectionParameters(host="localhost", credentials=credentials))
channel = connection.channel()

# Declare a queue named 'hello'
channel.queue_declare(queue="hello")

count = 0
while True:
    try:
        count += 1
        # Generate a fake user data and convert it to JSON format
        user = json.dumps({"id": count, "name": fake.name(), "address": fake.address(), "created_at": fake.year()})

        # Publish the user data to the 'hello' queue
        channel.basic_publish(exchange="", routing_key="hello", body=user)
        print(f" [x] Sent {user}")

        # Wait for 1 second before sending the next message
        # time.sleep(1)
    except KeyboardInterrupt:
        # Handle the interruption gracefully
        print("Interrupted")
        break
