#!/usr/bin/env python
import pika, sys, os, time


AUTO_ACK = False


def main():
    # Set the RabbitMQ server credentials
    credentials = pika.PlainCredentials("admin", "1234")

    # Establish a connection to the RabbitMQ server
    connection = pika.BlockingConnection(pika.ConnectionParameters(host="localhost", credentials=credentials))
    channel = connection.channel()

    # Declare a queue named 'hello'
    channel.queue_declare(queue="hello")

    # Define a callback function to process messages from the queue
    def callback(ch, method, properties, body):
        # read body as string
        body = body.decode("utf-8")
        # Print the received message
        print(f" [x] Received {body}")
        # time.sleep(1)
        if AUTO_ACK is False:
            ch.basic_ack(delivery_tag=method.delivery_tag)

    # Set up subscription on the queue 'hello' with the callback function
    channel.basic_consume(queue="hello", on_message_callback=callback, auto_ack=AUTO_ACK)

    print(" [*] Waiting for messages. To exit press CTRL+C")
    # Start consuming messages from the queue
    channel.start_consuming()


if __name__ == "__main__":
    try:
        main()
    except KeyboardInterrupt:
        print("Interrupted")
        try:
            sys.exit(0)
        except SystemExit:
            os._exit(0)
