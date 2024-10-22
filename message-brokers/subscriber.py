#!/usr/bin/env python
import pika, sys, os, time


def main():
    # Set the RabbitMQ server credentials
    credentials = pika.PlainCredentials("admin", "1234")

    # Establish a connection to RabbitMQ server
    connection = pika.BlockingConnection(pika.ConnectionParameters(host="localhost", credentials=credentials))
    channel = connection.channel()

    # Declare an exchange named 'logs' of type 'fanout'
    channel.exchange_declare(exchange="logs", exchange_type="fanout")

    # Declare a temporary queue with a random name, exclusive to this connection
    result = channel.queue_declare(queue="", exclusive=True)
    queue_name = result.method.queue

    # Bind the queue to the 'logs' exchange
    channel.queue_bind(exchange="logs", queue=queue_name)

    # Define a callback function to process messages from the queue
    def callback(ch, method, properties, body):
        # read body as string
        body = body.decode("utf-8")
        # Print the received message
        print(f" [x] Received {body}")
        time.sleep(1)
        # if AUTO_ACK is False:
        #     ch.basic_ack(delivery_tag=method.delivery_tag)

    # Start consuming messages from the queue with the callback function
    channel.basic_consume(queue=queue_name, on_message_callback=callback, auto_ack=True)

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
