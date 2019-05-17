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

	"github.com/goph/emperror"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/banzaicloud/allspark/internal/platform/healthcheck"
	"github.com/banzaicloud/allspark/internal/platform/log"
	"github.com/banzaicloud/allspark/internal/server"
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

	// Server
	Server server.Config `mapstructure:"server"`
}

// Validate validates the configuration
func (c Config) Validate() (Config, error) {
	config, err := c.Log.Validate()
	if err != nil {
		return c, emperror.Wrap(err, "could not validate log config")
	}
	c.Log = config

	config2, err := c.Healthcheck.Validate()
	if err != nil {
		return c, emperror.Wrap(err, "could not validate healthcheck config")
	}
	c.Healthcheck = config2

	config3, err := c.Server.Validate()
	if err != nil {
		return c, emperror.Wrap(err, "could not validate server config")
	}
	c.Server = config3

	return c, nil
}

func configure() {
	setupViper(viper.GetViper(), pflag.CommandLine)
	pflag.Parse()

	err := viper.ReadInConfig()
	if _, ok := err.(viper.ConfigFileNotFoundError); err != nil && !ok {
		panic(emperror.Wrap(err, "failed to read configuration"))
	}
	bindEnvs(configuration)
	err = viper.Unmarshal(&configuration)
	if err != nil {
		panic(emperror.Wrap(err, "failed to unmarshal configuration"))
	}

	configuration, err = configuration.Validate()
	if err != nil {
		panic(emperror.Wrap(err, "cloud not validate configuration"))
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
				panic(emperror.Wrap(err, "could not bind env variable"))
			}
		}
	}
}
