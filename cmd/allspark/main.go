// Copyright © 2019 Banzai Cloud
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"fmt"
	"os"
	"sync"

	"emperror.dev/emperror"
	"emperror.dev/errors"
	"github.com/banzaicloud/allspark/internal/kafka"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	yaml "gopkg.in/yaml.v2"

	"github.com/banzaicloud/allspark/internal/grpcserver"
	"github.com/banzaicloud/allspark/internal/httpserver"
	"github.com/banzaicloud/allspark/internal/platform/errorhandler"
	"github.com/banzaicloud/allspark/internal/platform/healthcheck"
	"github.com/banzaicloud/allspark/internal/platform/log"
	"github.com/banzaicloud/allspark/internal/request"
	"github.com/banzaicloud/allspark/internal/sql"
	"github.com/banzaicloud/allspark/internal/tcpserver"
	"github.com/banzaicloud/allspark/internal/workload"
)

// nolint: gochecknoinits
func init() {
	pflag.Bool("version", false, "Show version information")
	pflag.Bool("dump-config", false, "Dump configuration to the console")
}

func main() {
	// Loads and validates configuration
	configure()

	// Show version if asked for
	if viper.GetBool("version") {
		fmt.Printf("%s version %s (%s) built on %s\n", FriendlyServiceName, version, commitHash, buildDate)
		os.Exit(0)
	}

	// Dump config if asked for
	if viper.GetBool("dump-config") {
		c := viper.AllSettings()
		y, err := yaml.Marshal(c)
		if err != nil {
			panic(errors.WrapIf(err, "failed to dump configuration"))
		}
		fmt.Print(string(y))
		os.Exit(0)
	}

	// Create logger
	logger := log.NewLogger(configuration.Log)

	// Create error handler
	errorHandler := errorhandler.ErrorHandler(logger)
	defer emperror.HandleRecover(errorHandler)

	logger.Infof("starting %s", FriendlyServiceName)

	// Starts health check HTTP server
	go func() {
		healthcheck.New(configuration.Healthcheck, logger, errorHandler)
	}()

	var err error
	var sqlClient *sql.Client
	sqlQuery := viper.GetString("sql_query")
	sqlDSN := viper.GetString("sql_dsn")
	sqlQueryRepeatCount := viper.GetInt("sql_query_repeat_count")
	sqlQueryRepeatCountMax := viper.GetInt("sql_query_repeat_count_max")
	if sqlDSN != "" && sqlQuery != "" {
		sqlClient, err = sql.NewClient(sqlDSN, sqlQuery, sqlQueryRepeatCount, sqlQueryRepeatCountMax)
		if err != nil {
			panic(err)
		}
		logger.WithFields(log.Fields{
			"driver":              sqlClient.GetDriver(),
			"query":               sqlQuery,
			"queryRepeatCount":    sqlQueryRepeatCount,
			"queryRepeatCountMax": sqlQueryRepeatCountMax,
		}).Info("SQL client initialized")
	}

	requests, err := request.CreateRequestsFromStringSlice(viper.GetStringSlice("requests"), logger.WithField("server", "any"))
	if err != nil {
		panic(err)
	}

	var wl workload.Workload
	switch viper.GetString("workload") {
	case "Echo":
		str := viper.GetString("ECHO_STR")
		wl = workload.NewEchoWorkload(str, logger)
	case "PI":
		count := viper.GetInt("PI_COUNT")
		if count < 1 {
			count = 50000
		}
		wl = workload.NewPIWorkload(uint(count), logger)
	case "AIRPORTS":
		wl = workload.NewAirportWorkload(logger)
	}

	kafkaBootstrapServer := viper.GetString("kafka_bootstrap_server")

	// Kafka Consumer
	var kafkaConsumerConfig *kafka.Consumer
	kafkaConsumerTopic := viper.GetString("kafka_consumer")
	kafkaConsumerGroup := viper.GetString("kafka_consumer_group")

	if kafkaBootstrapServer != "" && kafkaConsumerTopic != "" {
		kafkaConsumerConfig = kafka.NewConsumer(kafkaBootstrapServer, kafkaConsumerTopic, kafkaConsumerGroup, logger)
		logger.WithFields(log.Fields{
			"bootstrapServer": kafkaBootstrapServer,
			"consumerTopic":   kafkaConsumerTopic,
			"consumerGroup":   kafkaConsumerGroup,
		}).Info("Kafka consumer initialized")
	}

	// Kafka Producer
	var kafkaProducerConfig *kafka.Producer
	kafkaProducerTopic := viper.GetString("kafka_producer")

	if kafkaBootstrapServer != "" && kafkaProducerTopic != "" {
		kafkaProducerConfig, err = kafka.NewProducer(kafkaBootstrapServer, kafkaProducerTopic, wl, logger)
		if err != nil {
			panic(err)
		}
		logger.WithFields(log.Fields{
			"bootstrapServer": kafkaBootstrapServer,
			"producerTopic":   kafkaProducerTopic,
			"workload":        wl.GetName(),
		}).Info("Kafka producer initialized")
	}

	var wg sync.WaitGroup

	// HTTP server
	wg.Add(1)
	go func() {
		defer wg.Done()
		srv := httpserver.New(configuration.HTTPServer, logger, errorHandler)
		if wl != nil {
			srv.SetWorkload(wl)
		}

		httpRequests, err := request.CreateRequestsFromStringSlice(viper.GetStringSlice("httpRequests"), logger.WithField("server", "http"))
		if err != nil {
			panic(err)
		}
		if len(httpRequests) == 0 {
			httpRequests = requests
		}

		srv.SetRequests(httpRequests)
		srv.SetSQLClient(sqlClient)
		srv.Run()
	}()

	// GRPC server
	wg.Add(1)
	go func() {
		defer wg.Done()
		srv := grpcserver.New(configuration.GRPCServer, logger, errorHandler)
		if wl != nil {
			srv.SetWorkload(wl)
		}

		grpcRequests, err := request.CreateRequestsFromStringSlice(viper.GetStringSlice("grpcRequests"), logger.WithField("server", "grpc"))
		if err != nil {
			panic(err)
		}
		if len(grpcRequests) == 0 {
			grpcRequests = requests
		}

		srv.SetRequests(grpcRequests)
		srv.SetSQLClient(sqlClient)
		srv.Run()
	}()

	// TCP server
	wg.Add(1)
	go func() {
		defer wg.Done()
		srv := tcpserver.New(configuration.TCPServer, logger, errorHandler)
		if wl != nil {
			srv.SetWorkload(wl)
		}

		tcpRequests, err := request.CreateRequestsFromStringSlice(viper.GetStringSlice("tcpRequests"), logger.WithField("server", "tcp"))
		if err != nil {
			panic(err)
		}
		if len(tcpRequests) == 0 {
			tcpRequests = requests
		}

		srv.SetRequests(tcpRequests)
		srv.SetSQLClient(sqlClient)
		srv.Run()
	}()

	if kafkaProducerConfig != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()

			kafkaProducerConfig.Produce(context.Background())
		}()
	}

	if kafkaConsumerConfig != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()

			kafkaConsumerConfig.Consume(context.Background())
		}()
	}

	wg.Wait()
}
