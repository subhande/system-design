import pika, json, time

from faker import Faker

fake = Faker()

connection = pika.BlockingConnection(pika.ConnectionParameters(host="localhost"))
channel = connection.channel()

channel.queue_declare(queue="hello")
count = 0
while True:
    try:
        count += 1
        user = json.dumps({"id": count, "name": fake.name(), "address": fake.address(), "created_at": fake.year()})
        channel.basic_publish(exchange="", routing_key="hello", body=user)
        print(f" [x] Sent {user}")
        # time.sleep(0.5)
    except KeyboardInterrupt:
        print("Interrupted")
        break
connection.close()
