# Message Brokers

## RabbitMQ

RabbitMQ is a message broker: it accepts and forwards messages. You can think about it as a post office: when you put the mail that you want posting in a post box, you can be sure that Mr. or Ms. Mailperson will eventually deliver the mail to your recipient. In this analogy, RabbitMQ is a post box, a post office and a postman.



## Functionality Test
- [x] Basic Producer and Consumer: Basic message sending and receiving
- [x] Work Queues: Distributing tasks among workers (Round-robin dispatching)
- [x] Publish/Subscribe: Broadcasting messages to multiple consumers
- [ ] Routing
- [ ] Topics
- [ ] RPC


## INSTALLATION

```bash
pip install -r requirements.txt
```

```bash
docker run -d -it --rm --name rabbitmq -e RABBITMQ_DEFAULT_USER=admin -e RABBITMQ_DEFAULT_PASS=1234 -p 5672:5672 -p 15672:15672 rabbitmq:3.13-management
```




# Management console
http://localhost:15672/