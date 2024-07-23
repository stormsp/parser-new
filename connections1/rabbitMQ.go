package PKG

import (
	"context"
	"encoding/json"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"sync"
	"time"
)

var ConnectToRabit bool ///Смотрим подключились ли к Rabbit
var InputMap SafeMap
var OutputMap SafeMap

type Out struct {
	MEK_Address int
	Raper       string
	Value       float32
	TypeParam   string
	OldValue    float32
	Reliability bool
	TimeOld     time.Time
	Time        time.Time
}

type OutToRabbitMQ struct {
	MEK_Address int
	Raper       string
	Value       float32
	TypeParam   string
	Reliability bool
	Time        time.Time
}

type SafeMap struct {
	Mu  sync.Mutex
	Out map[string]Out
}

var CONNECTRABBITMIB = "amqp://admin:admin@127.0.0.1:5672/"
var CONNECTRABBITPC = "amqp://guest:guest@192.168.1.200:5672/"

var ConnRabbitMQPublish *amqp.Connection
var ConnRabbitMQConsume *amqp.Connection
var NameAlg = "ButtonALG"

type SafeChange struct {
	Mu      sync.Mutex
	Changed bool
}

// InitializeRabbitMQConn создает  соединение для rabbit MQ «один процесс — одно соединение»
func InitializeRabbitMQConn(forWhat string) {
	c := make(chan *amqp.Error)
	go func() {
		err := <-c
		fmt.Println("Переподключение к Очереди: " + err.Error())
		InitializeRabbitMQConn(forWhat)
	}()

	conn, err := amqp.Dial(CONNECTRABBITMIB)
	//conn, err := amqp.Dial(CONNECTRABBITPC)
	if err != nil {
		fmt.Println("Не могу подключиться к Rabbit для ", forWhat, err)
	}
	conn.NotifyClose(c)
	if forWhat == "Publish" {
		ConnRabbitMQPublish = conn
	} else if forWhat == "Consume" {
		ConnRabbitMQConsume = conn
	}

}

// SendToRabbitMQ отправка Структуры в очередь по названию (для мэк)
func SendToRabbitMQ(OutputMap *SafeMap) {
	for {
		OutputMap.Mu.Lock()
		output := OutputMap.Out
		var outToRabbit = make([]OutToRabbitMQ, 0)
		for key, _ := range output {
			value := output[key]
			if value.TimeOld != value.Time {
				outToRabbit = append(outToRabbit, OutToRabbitMQ{value.MEK_Address, value.Raper, value.Value, value.TypeParam, value.Reliability, value.Time})
				outVal, exist := OutputMap.Out[key]
				if exist {
					outVal.TimeOld = outVal.Time
					OutputMap.Out[key] = outVal
				}
			}
		}
		OutputMap.Mu.Unlock()
		if len(outToRabbit) > 0 {
			body, err := json.Marshal(outToRabbit)
			if err != nil {
				fmt.Println("Ошибка При формировании JSON ", err)
			}
			ch, err := ConnRabbitMQPublish.Channel()
			if err != nil {
				fmt.Println("Ошибка открытия канала RabbitMQ ", err)
			}
			args := amqp.Table{
				"x-max-length": 1,
				"x-overflow":   "reject-publish",
			}
			q, err := ch.QueueDeclare(
				NameAlg+"Out", // name
				false,         // durable
				false,         // delete when unused
				false,         // exclusive
				false,         // no-wait
				args,          // arguments
			)
			if err != nil {
				fmt.Println("Failed to declare a queue ", err)

			}
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			err = ch.PublishWithContext(ctx,
				"",     // exchange
				q.Name, // routing key
				false,  // mandatory
				false,  // immediate
				amqp.Publishing{
					ContentType: "application/json",
					Body:        body,
				})
			if err != nil {
				fmt.Println("Ошибка отправки в очередь", err)
			}
			cancel()
			ch.Close()
		}
		time.Sleep(100 * time.Millisecond)
	}
}

// ConsumeFromRabbitMq получаем сообщения rabbit по названию очереди
func ConsumeFromRabbitMq(Out *SafeMap) {
	//Conn := ConnRabbitMQConsume
	ch, err := ConnRabbitMQConsume.Channel()
	if err != nil {
		fmt.Println("Ошибка открытия канала RabbitMQ ", err)
	}

	defer ch.Close()
	args := amqp.Table{
		"x-max-length": 1,
		"x-overflow":   "reject-publish",
	}
	q, err := ch.QueueDeclare(
		NameAlg, // name
		false,   // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		args,    // arguments
	)
	if err != nil {
		fmt.Println("Consumer Ошибка декларирования очереди RabbitMQ ", NameAlg+"Out", err)
	}

	err = ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		fmt.Println("Consumer Ошибка Qos RabbitMQ ", err)
	}

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		args,   // args
	)
	if err != nil {
		fmt.Println("Consumer Ошибка создания Consumer ", err)
	}

	var forever chan struct{}

	if err == nil {
		go MessageHandler(msgs, Out)
	}
	fmt.Println(" [*] Waiting for messages.")
	<-forever

}

// MessageHandler записывает изменения пришедшие с очереди мека в общую выходную структуру
func MessageHandler(msgs <-chan amqp.Delivery, OutArr *SafeMap) {
	var data []OutToRabbitMQ
	for d := range msgs {
		err := json.Unmarshal(d.Body, &data)
		if err != nil {
			fmt.Println("Ошибка разбора JSON:", err)
			continue
		}
		//  ************ ЗАПИСЬ В ОБЩУЮ СТРУКТУРУ**********
		OutArr.Mu.Lock()
		for _, inputVal := range data {
			ConnectToRabit = true
			outVal, exist := OutArr.Out[inputVal.Raper]
			if exist {
				outVal.Value = inputVal.Value
				outVal.Time = inputVal.Time
				OutArr.Out[inputVal.Raper] = outVal
			} else {
				fmt.Println("Добавляю в массив: ", inputVal.Raper)
				OutArr.Out[inputVal.Raper] = Out{
					Value:       inputVal.Value,
					Time:        inputVal.Time,
					Raper:       inputVal.Raper,
					MEK_Address: inputVal.MEK_Address,
					TypeParam:   inputVal.TypeParam,
				}
			}
		}
		OutArr.Mu.Unlock()
		d.Ack(false)
	}
}
func DeclareArrays() {
	InputMap.Out = make(map[string]Out)
	OutputMap.Out = make(map[string]Out)
}
func DeclareRabbit() {
	InitializeRabbitMQConn("Publish")
	InitializeRabbitMQConn("Consume")
}
func UpdateVal(tag string, val float32, Reliability bool) {
	OutputMap.Mu.Lock()
	outVal, exist := OutputMap.Out[tag]
	if exist {
		if outVal.Value == val {
			OutputMap.Mu.Unlock()
			return
		}
		outVal.Value = val
		outVal.TimeOld = outVal.Time
		outVal.Time = time.Now()
		outVal.Reliability = Reliability
		OutputMap.Out[tag] = outVal
	} else {
		OutputMap.Out[tag] = Out{
			Value:       val,
			Reliability: Reliability,
			Raper:       tag,
			Time:        time.Now(),
		}
	}
	OutputMap.Mu.Unlock()
	return
}
