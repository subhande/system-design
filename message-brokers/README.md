

## INSTALLATION

```bash
docker run -d -it --rm --name rabbitmq -p 5672:5672 -p 15672:15672 rabbitmq:3.13-management


docker run -d -it --rm --name rabbitmq -e RABBITMQ_DEFAULT_USER=admin -e RABBITMQ_DEFAULT_PASS=1234 -p 5672:5672 -p 15672:15672 rabbitmq:3.13-management

pip install pika

# Management console
http://localhost:15672/



```