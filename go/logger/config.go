package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"os/user"

	"github.com/spf13/viper"
)

var log Logger

func defaultLoggerCfg() Configuration {
	return Configuration{
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
	envConfig, _ := strconv.ParseBool(os.Getenv("PRLOG_ENVCONFIG"))
	if !envConfig {
		fmt.Printf("PRLOG_ENVCONFIG not true\n")
		return
	}

	config := new(Configuration)
	EnvConfigure(defaultLoggerCfg(), config, "PRLOG")
	// fmt.Printf("config %+v\n\n", config)

	if config.EnableKafka {
		kafkaCfg := new(ProducerConfiguration)
		EnvConfigure(defaultKafkaCfg(), kafkaCfg, "PRKAFKA")
		// fmt.Printf("kafkaCfg %+v\n\n", kafkaCfg)

		config.KafkaProducerCfg = *kafkaCfg
		// fmt.Printf("config %+v\n", config)
	}

	var err error
	logtype := os.Getenv("PRLOG_LOGTYPE")
	if logtype == "Zap" {
		log, err = NewLogger(*config, Zap)
	} else {
		log, err = NewLogger(*config, Logrus)
	}
	if err != nil {
		fmt.Printf("Could not instantiate %s logger %s", logtype, err.Error())
	}
}

func EnvConfigure(defaultCfg interface{}, config interface{}, prefix string) {
	var defaultMap map[string]interface{}
	defaultJson, _ := json.Marshal(defaultCfg)
	json.Unmarshal(defaultJson, &defaultMap)
	// fmt.Printf("defaultMap %+v\n\n", defaultMap)

	v := viper.New()
	for key, value := range defaultMap {
		// fmt.Printf("key %+v value %+v\n", key, value)
		v.SetDefault(key, value)
	}
	v.SetEnvPrefix(prefix)
	// v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// fmt.Printf("\nbefore %+v\n\n", config)
	if err := v.Unmarshal(config); err != nil {
		fmt.Println(err)
	}
	// fmt.Printf("after %+v\n\n", config)
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
