package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"os/user"

	"github.com/spf13/viper"
)

// Supported auto config environment names
const (
	LogAutoCfgEnvName         string = "PRLOG_AUTOCFG"
	KafkaAutoCfgEnvName              = "PRKAFKA_AUTOCFG"
	CloudEventsAutoCfgEnvName        = "PRCE_AUTOCFG"
)

// Supported auto configuration types
const (
	EnvConfig  string = "env"
	FileConfig        = "file"
	BothConfig        = "both"
)

// Supported environment name prefixes
const (
	LogEnvPrefix         string = "PRLOG"
	KafkaEnvPrefix              = "PRKAFKA"
	CloudEventsEnvPrefix        = "PRCE"
)

// Supported config file names
const (
	LogFileName         string = "pr_log_config"
	KafkaFileName              = "pr_kafka_config"
	CloudEventsFileName        = "pr_ce_config"
)

// logger global for go log pkg emulation
var logger Logger

// DefaultLogCfg returns default log configuration
func DefaultLogCfg() Configuration {
	return Configuration{
		LogPackage:        LogrusType,
		LogLevel:          InfoType,
		EnableTimeStamps:  true,
		EnableColorLevels: true,
		EnableCloudEvents: true,
		EnableKafka:       false,
		KafkaFormat:       CEFormat,
		EnableConsole:     false,
		ConsoleFormat:     TextFormat,
		EnableFile:        true,
		FileFormat:        JSONFormat,
		FileLocation:      "pavedroad.log",
	}
}

// DefaultKafkaCfg returns default kafka configuration
func DefaultKafkaCfg() ProducerConfiguration {
	var username string
	user, err := user.Current()
	if err != nil {
		username = "username"
	} else {
		username = user.Username
	}
	return ProducerConfiguration{
		Brokers:     []string{"localhost:9092"},
		Topic:       "logs",
		Partition:   RandomPartition,
		Key:         FixedKey,
		KeyName:     username,
		Compression: CompressionSnappy,
		AckWait:     WaitForLocal,
		FlushFreq:   500, // milliseconds
		EnableTLS:   false,
		EnableDebug: false,
	}
}

// DefaultCloudEventsCfg returns default cloudevents configuration
func DefaultCloudEventsCfg() CloudEventsConfiguration {
	return CloudEventsConfiguration{
		Source:      "http://github.com/pavedroad-io/core/go/logger",
		SpecVersion: "1.0",
		Type:        "io.pavedroad.cloudevents.log",
		ID:          HMAC,
	}
}

// init called on package import to configure logger via environment
func init() {
	var err error
	config := new(Configuration)
	err = EnvConfigure(DefaultLogCfg(), config, os.Getenv(LogAutoCfgEnvName),
		LogFileName, LogEnvPrefix)
	if err != nil {
		fmt.Printf("Could not create logger configuration %s:", err.Error())
		os.Exit(1)
	}

	kafkaConfig := new(ProducerConfiguration)
	err = EnvConfigure(DefaultKafkaCfg(), kafkaConfig,
		os.Getenv(KafkaAutoCfgEnvName), KafkaFileName, KafkaEnvPrefix)
	if err == nil {
		config.KafkaProducerCfg = *kafkaConfig
	} else {
		fmt.Printf("Could not create kafka configuration %s:", err.Error())
		if config.EnableKafka {
			os.Exit(1)
		}
	}

	ceConfig := new(CloudEventsConfiguration)
	err = EnvConfigure(DefaultCloudEventsCfg(), ceConfig,
		os.Getenv(CloudEventsAutoCfgEnvName), CloudEventsFileName, CloudEventsEnvPrefix)
	if err == nil {
		config.CloudEventsCfg = *ceConfig
	} else {
		fmt.Printf("Could not create cloudevents configuration %s:", err.Error())
		if config.EnableCloudEvents {
			os.Exit(1)
		}
	}

	if logger, err = NewLogger(*config); err != nil {
		fmt.Printf("Could not instantiate %s logger package: %s",
			config.LogPackage, err.Error())
		os.Exit(1)
	}
}

// EnvConfigure fills config from defaults, config file and environment
func EnvConfigure(defaultCfg interface{}, config interface{}, auto string,
	filename string, prefix string) error {

	var defaultMap map[string]interface{}
	defaultJSON, err := json.Marshal(defaultCfg)
	if err != nil {
		return err
	}

	err = json.Unmarshal(defaultJSON, &defaultMap)
	if err != nil {
		return err
	}

	v := viper.New()
	for key, value := range defaultMap {
		v.SetDefault(key, value)
	}
	if auto == EnvConfig || auto == BothConfig {
		v.SetEnvPrefix(prefix)
		v.AutomaticEnv()
	}

	if auto == FileConfig || auto == BothConfig {
		v.SetConfigName(filename)
		v.AddConfigPath(".")
		v.AddConfigPath("$HOME")
		v.AddConfigPath("$HOME/.pavedroad.d")
		if err := v.ReadInConfig(); err != nil {
			return err
		}
	}

	if err := v.Unmarshal(config); err != nil {
		return err
	}
	return nil
}

// Print emulates function from go log pkg
func Print(args ...interface{}) {
	logger.Info(args...)
}

// Printf emulates function from go log pkg
func Printf(format string, args ...interface{}) {
	logger.Infof(format, args...)
}

// Println emulates function from go log pkg
func Println(args ...interface{}) {
	logger.Info(strings.TrimRight(fmt.Sprintln(args...), "\n"))
}

// Debug emulates function from go log pkg
func Debug(args ...interface{}) {
	logger.Debug(args...)
}

// Debugf emulates function from go log pkg
func Debugf(format string, args ...interface{}) {
	logger.Debugf(format, args...)
}

// Debugln emulates function from go log pkg
func Debugln(args ...interface{}) {
	logger.Debug(strings.TrimRight(fmt.Sprintln(args...), "\n"))
}

// Info emulates function from go log pkg
func Info(args ...interface{}) {
	logger.Info(args...)
}

// Infof emulates function from go log pkg
func Infof(format string, args ...interface{}) {
	logger.Infof(format, args...)
}

// Infoln emulates function from go log pkg
func Infoln(args ...interface{}) {
	logger.Info(strings.TrimRight(fmt.Sprintln(args...), "\n"))
}

// Warn emulates function from go log pkg
func Warn(args ...interface{}) {
	logger.Warn(args...)
}

// Warnf emulates function from go log pkg
func Warnf(format string, args ...interface{}) {
	logger.Warnf(format, args...)
}

// Warnln emulates function from go log pkg
func Warnln(args ...interface{}) {
	logger.Warn(strings.TrimRight(fmt.Sprintln(args...), "\n"))
}

// Error emulates function from go log pkg
func Error(args ...interface{}) {
	logger.Error(args...)
}

// Errorf emulates function from go log pkg
func Errorf(format string, args ...interface{}) {
	logger.Errorf(format, args...)
}

// Errorln emulates function from go log pkg
func Errorln(args ...interface{}) {
	logger.Error(strings.TrimRight(fmt.Sprintln(args...), "\n"))
}

// Fatal emulates function from go log pkg
func Fatal(args ...interface{}) {
	logger.Fatal(args...)
}

// Fatalf emulates function from go log pkg
func Fatalf(format string, args ...interface{}) {
	logger.Fatalf(format, args...)
}

// Fatalln emulates function from go log pkg
func Fatalln(args ...interface{}) {
	logger.Fatal(strings.TrimRight(fmt.Sprintln(args...), "\n"))
}

// Panic emulates function from go log pkg
func Panic(args ...interface{}) {
	logger.Panic(args...)
}

// Panicf emulates function from go log pkg
func Panicf(format string, args ...interface{}) {
	logger.Fatalf(format, args...)
}

// Panicln emulates function from go log pkg
func Panicln(args ...interface{}) {
	logger.Panic(strings.TrimRight(fmt.Sprintln(args...), "\n"))
}
