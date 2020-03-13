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
	LogAutoCfgEnvName   string = "PRLOG_AUTOCFG"
	KafkaAutoCfgEnvName        = "PRKAFKA_AUTOCFG"
)

// Supported auto configuration types
const (
	EnvConfig  string = "env"
	FileConfig        = "file"
	BothConfig        = "both"
)

// Supported environment name prefixes
const (
	LogEnvPrefix   string = "PRLOG"
	KafkaEnvPrefix        = "PRKAFKA"
)

// Supported config file names
const (
	LogFileName   string = "pr_log_config"
	KafkaFileName        = "pr_kafka_config"
)

var log Logger

func defaultLogCfg() Configuration {
	return Configuration{
		LogPackage:           LogrusType,
		LogLevel:             InfoType,
		EnableCloudEvents:    true,
		EnableKafka:          false,
		KafkaFormat:          CEFormat,
		EnableConsole:        true,
		ConsoleFormat:        TextFormat,
		ConsoleLevelTruncate: true,
		EnableFile:           false,
		FileFormat:           JSONFormat,
		FileLevelTruncate:    false,
		FileLocation:         "pavedroad.log",
	}
}

func defaultKafkaCfg() ProducerConfiguration {
	user, _ := user.Current()
	return ProducerConfiguration{
		Brokers:       []string{"localhost:9092"},
		Topic:         "logs",
		Partition:     RandomPartition,
		Key:           FixedKey,
		KeyName:       user.Username,
		CloudeventsID: HMAC,
		Compression:   CompressionSnappy,
		AckWait:       WaitForLocal,
		FlushFreq:     500, // milliseconds
		EnableTLS:     false,
	}
}

func init() {
	var err error
	config := new(Configuration)
	err = EnvConfigure(defaultLogCfg(), config, os.Getenv(LogAutoCfgEnvName),
		LogFileName, LogEnvPrefix)
	// fmt.Printf("config %+v\n\n", config)
	if err != nil {
		fmt.Printf("Could not create logger configuration %s:", err.Error())
		os.Exit(1)
	}
	kafkaConfig := new(ProducerConfiguration)
	err = EnvConfigure(defaultKafkaCfg(), kafkaConfig, os.Getenv(KafkaAutoCfgEnvName),
		KafkaFileName, KafkaEnvPrefix)
	// fmt.Printf("config %+v\n\n", config)
	if err != nil {
		fmt.Printf("Could not create kafka configuration %s:", err.Error())
		os.Exit(1)
	}
	config.KafkaProducerCfg = *kafkaConfig

	if log, err = NewLogger(*config); err != nil {
		fmt.Printf("Could not instantiate %s logger package: %s",
			config.LogPackage, err.Error())
		os.Exit(1)
	}
}

func EnvConfigure(defaultCfg interface{}, config interface{}, auto string,
	filename string, prefix string) error {

	var defaultMap map[string]interface{}
	defaultJson, err := json.Marshal(defaultCfg)
	if err != nil {
		return err
	}
	json.Unmarshal(defaultJson, &defaultMap)
	// fmt.Printf("defaultMap %+v\n\n", defaultMap)

	v := viper.New()
	for key, value := range defaultMap {
		// fmt.Printf("key %+v value %+v\n", key, value)
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

	// fmt.Printf("\nbefore %+v\n\n", config)
	if err := v.Unmarshal(config); err != nil {
		return err
	}
	// fmt.Printf("after %+v\n\n", config)
	return nil
}

func Print(args ...interface{}) {
	log.Info(args...)
}

func Printf(format string, args ...interface{}) {
	log.Infof(format, args...)
}

func Println(args ...interface{}) {
	log.Info(strings.TrimRight(fmt.Sprintln(args...), "\n"))
}

func Debug(args ...interface{}) {
	log.Debug(args...)
}

func Debugf(format string, args ...interface{}) {
	log.Debugf(format, args...)
}

func Debugln(args ...interface{}) {
	log.Debug(strings.TrimRight(fmt.Sprintln(args...), "\n"))
}

func Info(args ...interface{}) {
	log.Info(args...)
}

func Infof(format string, args ...interface{}) {
	log.Infof(format, args...)
}

func Infoln(args ...interface{}) {
	log.Info(strings.TrimRight(fmt.Sprintln(args...), "\n"))
}

func Warn(args ...interface{}) {
	log.Warn(args...)
}

func Warnf(format string, args ...interface{}) {
	log.Warnf(format, args...)
}

func Warnln(args ...interface{}) {
	log.Warn(strings.TrimRight(fmt.Sprintln(args...), "\n"))
}

func Error(args ...interface{}) {
	log.Error(args...)
}

func Errorf(format string, args ...interface{}) {
	log.Errorf(format, args...)
}

func Errorln(args ...interface{}) {
	log.Error(strings.TrimRight(fmt.Sprintln(args...), "\n"))
}

func Fatal(args ...interface{}) {
	log.Fatal(args...)
}

func Fatalf(format string, args ...interface{}) {
	log.Fatalf(format, args...)
}

func Fatalln(args ...interface{}) {
	log.Fatal(strings.TrimRight(fmt.Sprintln(args...), "\n"))
}

func Panic(args ...interface{}) {
	log.Panic(args...)
}

func Panicf(format string, args ...interface{}) {
	log.Fatalf(format, args...)
}

func Panicln(args ...interface{}) {
	log.Panic(strings.TrimRight(fmt.Sprintln(args...), "\n"))
}
