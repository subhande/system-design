#!/usr/bin/env python
import pika, sys, os, time


def main():
    connection = pika.BlockingConnection(pika.ConnectionParameters(host="localhost"))
    channel = connection.channel()

    channel.queue_declare(queue="hello")

    def callback(ch, method, properties, body):
        # read body as string
        body = body.decode("utf-8")
        print(f" [x] Received {body}")
        # time.sleep(1)
        ch.basic_ack(delivery_tag=method.delivery_tag)

    channel.basic_consume(queue="hello", on_message_callback=callback, auto_ack=False)

    print(" [*] Waiting for messages. To exit press CTRL+C")
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
