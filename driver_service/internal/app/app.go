package app

import (
	"context"
	"driver_service/internal/config"
	"driver_service/internal/constants"
	"driver_service/internal/controller/http_handler"
	"driver_service/internal/controller/kafka_handler"
	"driver_service/internal/repository"
	"driver_service/internal/usecase"

	"flag"
	"fmt"
	"net"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

var (
	partitions = flag.Int("p", 1, "paritions numbers")
	replicas   = flag.Int("r", 1, "replicas numbers")
)

type AppServer struct {
	cfg           *config.Config
	server        *http.Server
	kafkaConsumer *kafka_handler.KafkaConsumer
	kafkaProducer *kafka.Writer
	logger        *zap.Logger
}

func NewAppServer(cfg *config.Config, logger *zap.Logger) *AppServer {
	address := fmt.Sprintf(":%d", cfg.Http.Port)

	mongo_new := repository.NewTripRepository()

	kafkaConsumer := kafka_handler.NewKafkaConsumer([]string{"127.0.0.1:29092"}, constants.DRIVER_SEARCH, mongo_new.Database)
	kafkaProducer := kafka_handler.NewKafkaProducer([]string{"127.0.0.1:29092", "127.0.0.1:39092", "127.0.0.1:49092"}, "FROM_DRIVER")

	a := &AppServer{
		cfg: cfg,
		server: &http.Server{
			Addr:    address,
			Handler: initApi(cfg, kafkaProducer, logger),
		},
		kafkaConsumer: kafkaConsumer,
		kafkaProducer: kafkaProducer,
		logger:        logger,
	}

	return a
}

func initApi(cfg *config.Config, writer *kafka.Writer, logger *zap.Logger) http.Handler {
	repo := repository.NewTripRepository()
	r := mux.NewRouter()
	uc := usecase.NewTripUsecase(repo)
	srvHandlers := http_handler.NewHttpHandler(uc, writer, logger)
	r.HandleFunc("/trips/{trip_id}/cancel", srvHandlers.CancelTrip)
	r.HandleFunc("/trips/{trip_id}/accept", srvHandlers.AcceptTrip)
	r.HandleFunc("/trips/{trip_id}/start", srvHandlers.StartTrip)
	r.HandleFunc("/trips/{trip_id}/end", srvHandlers.EndTrip)
	r.HandleFunc("/trips", srvHandlers.GetTrips)
	r.HandleFunc("/trips/{trip_id}", srvHandlers.GetTripByID)

	router := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		r.ServeHTTP(w, req)
	})

	return router

}

func InitKafkaTopic(topic_name string, port string, logger *zap.Logger) {

	ctx := context.Background()

	conn, err := kafka.DialContext(ctx, "tcp", fmt.Sprintf("127.0.0.1:%s", port))
	if err != nil {
		logger.Fatal("Failed to connect to Kafka", zap.Error(err))
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		logger.Fatal("Failed to get Kafka controller", zap.Error(err))
	}
	var controllerConn *kafka.Conn
	controllerConn, err = kafka.Dial("tcp", net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))
	if err != nil {
		logger.Fatal("Failed to connect to Kafka controller", zap.Error(err))
	}
	defer controllerConn.Close()

	topicConfigs := []kafka.TopicConfig{
		{
			Topic:             topic_name,
			NumPartitions:     *partitions,
			ReplicationFactor: *replicas,
			ConfigEntries: []kafka.ConfigEntry{
				// {ConfigName: "min.insync.replicas", ConfigValue: "2"},
				{ConfigName: "segment.bytes", ConfigValue: "2097152"},
				// {ConfigName: "retention.bytes", ConfigValue: "3145728"},
			},
		},
	}

	err = controllerConn.CreateTopics(topicConfigs...)
	if err != nil {
		logger.Fatal("Failed to create Kafka topic", zap.Error(err))
	}

	logger.Info("Successfully initialized Kafka topic")
}

func (a *AppServer) Run() {
	go func() {
		InitKafkaTopic("FROM_DRIVER", constants.PORT_TOPIC_DRIVER, a.logger)
		a.kafkaConsumer.Consume()
	}()

	go func() {
		err := a.server.ListenAndServe()
		if err != nil {
			a.logger.Error("Failed to start HTTP server", zap.Error(err))
		}
	}()
}

func (a *AppServer) Stop(ctx context.Context) {
	a.kafkaConsumer.Close()
	err := a.server.Shutdown(ctx)
	if err != nil {
		a.logger.Error("Error shutting down server", zap.Error(err))
	} else {
		a.logger.Info("Server shutdown successfully")
	}
}
