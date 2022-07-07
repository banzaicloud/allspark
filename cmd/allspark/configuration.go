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
	"fmt"
	"os"
	"reflect"
	"strings"

	"emperror.dev/errors"
	"github.com/banzaicloud/allspark/internal/kafka"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/banzaicloud/allspark/internal/grpcserver"
	"github.com/banzaicloud/allspark/internal/httpserver"
	"github.com/banzaicloud/allspark/internal/platform/healthcheck"
	"github.com/banzaicloud/allspark/internal/platform/log"
	"github.com/banzaicloud/allspark/internal/tcpserver"
)

// main configuration
var configuration Config

// Config holds any kind of configuration that comes from the outside world and
// is necessary for running the application
type Config struct {
	// Meaningful values are recommended (eg. production, development, staging, release/123, etc)
	Environment string `mapstucture:"environment"`

	// Turns on some debug functionality (eg. more verbose logs)
	Debug bool `mapstructure:"debug"`

	// Log configuration
	Log log.Config `mapstructure:"log"`

	// Healthcheck configuration
	Healthcheck healthcheck.Config `mapstructure:"healthcheck"`

	// HTTP server configuration
	HTTPServer httpserver.Config `mapstructure:"httpServer"`

	// GRPC server configuration
	GRPCServer grpcserver.Config `mapstructure:"grpcServer"`

	// TCP server configuration
	TCPServer tcpserver.Config `mapstructure:"tcpServer"`

	// Kafka server consumer configurations
	KafkaServer kafka.Consumer `mapstructure:"kafkaServer"`
}

// Validate validates the configuration
func (c Config) Validate() (Config, error) {
	logConfig, err := c.Log.Validate()
	if err != nil {
		return c, errors.WrapIf(err, "could not validate log config")
	}
	c.Log = logConfig

	healthCheckConfig, err := c.Healthcheck.Validate()
	if err != nil {
		return c, errors.WrapIf(err, "could not validate healthcheck config")
	}
	c.Healthcheck = healthCheckConfig

	serverConfig, err := c.HTTPServer.Validate()
	if err != nil {
		return c, errors.WrapIf(err, "could not validate HTTP server config")
	}
	c.HTTPServer = serverConfig

	grpcServerConfig, err := c.GRPCServer.Validate()
	if err != nil {
		return c, errors.WrapIf(err, "could not validate GRPC server config")
	}
	c.GRPCServer = grpcServerConfig

	tcpServerConfig, err := c.TCPServer.Validate()
	if err != nil {
		return c, errors.WrapIf(err, "could not validate TCP server config")
	}
	c.TCPServer = tcpServerConfig

	kafkaServerConfig, err := c.KafkaServer.Validate()
	if err != nil {
		return c, errors.WrapIf(err, "could not validate kafka server config")
	}
	c.KafkaServer = *kafkaServerConfig

	return c, nil
}

func configure() {
	setupViper(viper.GetViper(), pflag.CommandLine)
	pflag.Parse()

	err := viper.ReadInConfig()
	if _, ok := err.(viper.ConfigFileNotFoundError); err != nil && !ok {
		panic(errors.WrapIf(err, "failed to read configuration"))
	}
	bindEnvs(configuration)
	err = viper.Unmarshal(&configuration)
	if err != nil {
		panic(errors.WrapIf(err, "failed to unmarshal configuration"))
	}

	configuration, err = configuration.Validate()
	if err != nil {
		panic(errors.WrapIf(err, "cloud not validate configuration"))
	}
}

// setupViper configures some defaults in the Viper instance
func setupViper(v *viper.Viper, p *pflag.FlagSet) {
	v.AddConfigPath(".")
	v.AddConfigPath("./config")
	v.AddConfigPath("$HOME/config")
	p.Init(FriendlyServiceName, pflag.ExitOnError)
	pflag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", FriendlyServiceName)
		pflag.PrintDefaults()
	}
	v.BindPFlags(p) // nolint:errcheck

	v.SetEnvPrefix(ConfigEnvPrefix)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
}

func bindEnvs(iface interface{}, parts ...string) {
	ifv := reflect.ValueOf(iface)
	ift := reflect.TypeOf(iface)
	for i := 0; i < ift.NumField(); i++ {
		v := ifv.Field(i)
		t := ift.Field(i)
		tv, ok := t.Tag.Lookup("mapstructure")
		if !ok {
			continue
		}
		switch v.Kind() {
		case reflect.Struct:
			bindEnvs(v.Interface(), append(parts, tv)...)
		default:
			err := viper.BindEnv(strings.Join(append(parts, tv), "."))
			if err != nil {
				panic(errors.WrapIf(err, "could not bind env variable"))
			}
		}
	}
}
