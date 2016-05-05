# Distributed application with Golang
  * Tech
    * RabbitMQ
      * AMQP: advanced message queuing protocol - for message data format
      * Can be set up in a cluster for high available systems
      * Architecture
      * Install: https://www.rabbitmq.com/install-homebrew.html
      * Error:
        * "Can't set short node name!\n Please check your configuration\n"
        ```
        http://stackoverflow.com/questions/10767037/control-rabbitmq-name-not-sname
        $ sudo scutil --set YourHostName
        $ sudo vim /etc/hosts, change "127.0.0.1 localhost YourHostName"
        ```
      * Commands:
        ```
        $ rabbitmqctl status
        $ rabbitmqctl list_queues
        $ rabbitmqctl cluster_status
        $ rabbitmq-plugins list (for other message brokers)
        $ rabbitmq-plugins enable rabbitmq_management (http://localhost:15672/)
        $ rabbitmq-plugins disable rabbitmq_management
        ```
      * Golang support
        https://godoc.org/github.com/streadway/amqp, rabbitmq client with AMQP protocol
        Install: $ go get -u github.com/streadway/amqp
    * Postgres
  * Flow
    * Sensors keep publishing reading data to message queues
    * Consumers keep consuming messages and generate events (This event pattern allows data sources and consumers to be decoupled from each other in a highly concurrent system)
    ```
    $ go run src/powerplant/coordinator/executor/main.go (consuming messages)
    $ go run src/powerplant/sensors/executor/main.go (publishing messages)
    ```
