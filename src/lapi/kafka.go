package lapi

import (
	"log"
	"strings"

	"github.com/Shopify/sarama"
)

//func init() {
//sarama.Logger = log.New(os.Stdout, "[Sarama] ", log.LstdFlags)
//}

func Kconn(logserver string) sarama.SyncProducer {
	splitBrokers := strings.Split(logserver, ",")
	conf := sarama.NewConfig()
	conf.Producer.Retry.Max = 3
	conf.Producer.RequiredAcks = sarama.WaitForAll
	conf.Producer.Return.Successes = true
	conf.Metadata.Full = true
	conf.Version = sarama.V0_10_0_0
	conf.ClientID = "sasl_scram_client"
	conf.Metadata.Full = true

	syncProducer, err := sarama.NewSyncProducer(splitBrokers, conf)
	if err != nil {
		//logger.Fatalln("failed to create producer: ", err)
		log.Fatalln("failed to create producer: ", err)
	}
	return syncProducer
}

func Kafka(topic, message string, syncProducer sarama.SyncProducer) {

	partition, offset, err := syncProducer.SendMessage(&sarama.ProducerMessage{
		Topic: topic,
		//Key: sarama.StringEncoder("test"),
		Value: sarama.StringEncoder(message),
	})
	if err != nil {
		//logger.Fatalln("failed to send message to ", "topic", err)
		log.Fatalln("failed to send message to ", "topic", err)
		//logger.Printf("wrote message at partition: %d, offset: %d", partition, offset)
		log.Printf("wrote message at partition: %d, offset: %d", partition, offset)
	}
	//logger.Printf("wrote message at partition: %d, offset: %d", partition, offset)
	//_ = syncProducer.Close()
	//logger.Println("Bye now !")

}
