Hermes is a MQTT v5.0 compatible broker, so any compatible Java library can be used.

The widely used Eclipse Paho library can be imported into any Maven or Gradle project.

```xml

<repositories>
    <repository>
        <id>Eclipse Paho Repo</id>
        <url>https://repo.eclipse.org/content/repositories/paho-releases/</url>
    </repository>
</repositories>
```

```xml

<dependencies>
    <dependency>
        <groupId>org.eclipse.paho</groupId>
        <artifactId>org.eclipse.paho.mqttv5.client</artifactId>
        <version>1.2.5</version>
    </dependency>
</dependencies>
```

For Gradle repositories, use the below

```groovy
// Groovy script
repositories {
    maven {
        url "https://repo.eclipse.org/content/repositories/paho-releases/"
    }
}
```

```kotlin
// Kotlin script
repositories {
    maven {
        url = uri("https://repo.eclipse.org/content/repositories/paho-releases/")
    }
}
```

### Connecting to MQTT broker

```java
var persistence = new MemoryPersistence();
var client = new MqttClient(broker, clientID, persistence);
var connOpts = new MqttConnectionOptions();
client.connect(connOpts);
```

### Publishing messages

```java
// Ensure client has already connected
var content = "Hello World";
var topic = "my-topic";
var message = new MqttMessage(content.getBytes());
client.publish(topic,message);
```

### Subscribing to incoming messages

```java
// Do this before connecting
client.setCallback(new MqttCallback(){
    @Override
    public void messageArrived(String topic, MqttMessage message) throws Exception{
        // Do something awesome
    }
});
client.connect(connOpts);

// Provide a topic and Quality of Service (QoS)
client.subscribe("my-topic", 0);
```

## Spring Integration

Spring Integration provides inbound and outbound channel adapters to support the MQTT protocol.

The following dependencies can be used for Maven and Gradle respectively

```xml
<dependency>
    <groupId>org.springframework.integration</groupId>
    <artifactId>spring-integration-mqtt</artifactId>
    <version>5.4.2</version>
</dependency>
```

```groovy
compile "org.springframework.integration:spring-integration-mqtt:5.4.2"
```

### Inbound Channel Adapters
Inbound adapters allow Spring applications to subscribe to topics and respond to incoming MQTT messages.
```java
@Bean
public IntegrationFlow mqttInbound() {
    var broker = "tcp://localhost:1883";
    var clientID = "client-id";
    var topic = "my-topic";
    var adapter = new MqttPahoMessageDrivenChannelAdapter(broker, clientID, topic); 
    return IntegrationFlows.from(adapter).handle(m -> handleMsg(m)).get();
}

public void handleMsg(MqttMessage message) {
    // Do something awesome
}
```
### Outbound Channel Adapters
Inbound adapters allow Spring applications to publish MQTT messages onto topics.
```java
@Bean
public IntegrationFlow mqttOutboundFlow() {
    var broker = "tcp://localhost:1883";
    var clientID = "client-id";
    return f -> f.handle(new MqttPahoMessageHandler(broker, clientID));
}
```

More information regarding Spring MQTT integration can be found below 
on the [**Spring MQTT Support Homepage**](https://docs.spring.io/spring-integration/reference/html/mqtt.html#mqtt)


