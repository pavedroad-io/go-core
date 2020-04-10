package logger

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/user"
	"strings"

	"github.com/spf13/viper"
)

// Supported auto config environment names
const (
	LogAutoCfgEnvName         = "PRLOG_AUTOCFG"
	KafkaAutoCfgEnvName       = "PRKAFKA_AUTOCFG"
	CloudEventsAutoCfgEnvName = "PRCE_AUTOCFG"
	RotationAutoCfgEnvName    = "PRROT_AUTOCFG"
)

// Supported auto configuration types
const (
	EnvConfig  = "env"
	FileConfig = "file"
	BothConfig = "both"
)

// Supported environment name prefixes
const (
	LogEnvPrefix         = "PRLOG"
	KafkaEnvPrefix       = "PRKAFKA"
	CloudEventsEnvPrefix = "PRCE"
	RotationEnvPrefix    = "PRROT"
)

// Supported config file names
const (
	LogFileName         = "pr_log_config"
	KafkaFileName       = "pr_kafka_config"
	CloudEventsFileName = "pr_ce_config"
	RotationFileName    = "pr_rot_config"
)

// logger global for go log pkg emulation
var logger Logger

var defaultLoggerConfiguration = LoggerConfiguration{
	LogPackage:        ZapType,
	LogLevel:          InfoType,
	EnableTimeStamps:  true,
	EnableColorLevels: true,
	EnableCloudEvents: true,
	EnableKafka:       false,
	KafkaFormat:       CEFormat,
	EnableConsole:     false,
	ConsoleFormat:     TextFormat,
	ConsoleWriter:     Stdout,
	EnableFile:        true,
	FileFormat:        JSONFormat,
	FileLocation:      "pavedroad.log",
	EnableRotation:    false,
	EnableDebug:       true,
}

var defaultProducerConfiguration = ProducerConfiguration{
	Brokers:       []string{"localhost:9092"},
	Topic:         "logs",
	Partition:     RandomPartition,
	Key:           FixedKey,
	KeyName:       "username",
	Compression:   CompressionSnappy,
	AckWait:       WaitForLocal,
	ProdFlushFreq: 500, // milliseconds
	ProdRetryMax:  10,
	ProdRetryFreq: 100, // milliseconds
	MetaRetryMax:  10,
	MetaRetryFreq: 2000, // milliseconds
	EnableTLS:     false,
	EnableDebug:   false,
}

var defaultCloudEventsConfiguration = CloudEventsConfiguration{
	SetID:           CEHMAC,
	Source:          "http://github.com/pavedroad-io/go-core/logger",
	SpecVersion:     "1.0",
	Type:            "io.pavedroad.cloudevents.log",
	SetSubjectLevel: true,
}

var defaultRotationConfiguration = RotationConfiguration{
	Filename:   "pavedroad.log",
	MaxSize:    100, // megabytes
	MaxAge:     0,   // days, 0 = no expiration
	MaxBackups: 0,   // keep all
	LocalTime:  false,
	Compress:   false,
}

// DefaultLoggerCfg returns default log configuration
func DefaultLoggerCfg() LoggerConfiguration {
	return defaultLoggerConfiguration
}

// DefaultProducerCfg returns default kafka configuration
func DefaultProducerCfg() ProducerConfiguration {
	return defaultProducerConfiguration
}

// DefaultCloudEventsCfg returns default cloudevents configuration
func DefaultCloudEventsCfg() CloudEventsConfiguration {
	return defaultCloudEventsConfiguration
}

// DefaultRotationCfg returns default cloudevents configuration
func DefaultRotationCfg() RotationConfiguration {
	return defaultRotationConfiguration
}

// init called on package import to configure logger via environment
func init() {
	var err error

	user, err := user.Current()
	if err == nil {
		defaultProducerConfiguration.KeyName = user.Username
	}

	config := new(LoggerConfiguration)
	err = EnvConfigure(DefaultLoggerCfg(), config, os.Getenv(LogAutoCfgEnvName),
		LogFileName, LogEnvPrefix)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not create logger configuration: %s\n",
			err.Error())
		os.Exit(1)
	}

	kafkaConfig := new(ProducerConfiguration)
	err = EnvConfigure(DefaultProducerCfg(), kafkaConfig,
		os.Getenv(KafkaAutoCfgEnvName), KafkaFileName, KafkaEnvPrefix)
	if err == nil {
		config.KafkaProducerCfg = *kafkaConfig
	} else {
		fmt.Fprintf(os.Stderr, "Could not create kafka configuration: %s\n",
			err.Error())
		if config.EnableKafka {
			os.Exit(1)
		}
	}

	ceConfig := new(CloudEventsConfiguration)
	err = EnvConfigure(DefaultCloudEventsCfg(), ceConfig,
		os.Getenv(CloudEventsAutoCfgEnvName), CloudEventsFileName,
		CloudEventsEnvPrefix)
	if err == nil {
		config.CloudEventsCfg = *ceConfig
	} else {
		fmt.Fprintf(os.Stderr,
			"Could not create cloudevents configuration: %s\n", err.Error())
		if config.EnableCloudEvents {
			os.Exit(1)
		}
	}

	rotConfig := new(RotationConfiguration)
	err = EnvConfigure(DefaultRotationCfg(), rotConfig,
		os.Getenv(RotationAutoCfgEnvName), RotationFileName, RotationEnvPrefix)
	if err == nil {
		config.RotationCfg = *rotConfig
	} else {
		fmt.Fprintf(os.Stderr, "Could not create rotation configuration: %s\n",
			err.Error())
		if config.EnableRotation {
			os.Exit(1)
		}
	}

	if logger, err = NewLogger(*config); err != nil {
		fmt.Fprintf(os.Stderr, "Could not instantiate %s logger package: %s\n",
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

func checkConfig(config LoggerConfiguration) error {
	var errCount int

	checkLoggerConfig(config, &errCount)

	if config.EnableKafka {
		checkProducerConfig(config.KafkaProducerCfg, &errCount)
		if config.EnableCloudEvents {
			checkCETypes(config.CloudEventsCfg, &errCount)
		}
	}
	if config.EnableRotation {
		checkRotationConfig(config.RotationCfg, &errCount)
	}

	if errCount > 0 {
		return errors.New("Invalid configuration")
	}
	return nil
}

func checkLoggerConfig(lc LoggerConfiguration, errCount *int) {
	checkLoggerTypes(lc, errCount)

	if (lc.ConsoleFormat == CEFormat || lc.FileFormat == CEFormat ||
		lc.KafkaFormat == CEFormat) && !lc.EnableCloudEvents {
		fmt.Fprintf(os.Stderr, "CEFormat requires EnableCloudEvents\n")
		*errCount++
	}
}

func checkProducerConfig(pc ProducerConfiguration, errCount *int) {
	checkProducerTypes(pc, errCount)
	if pc.EnableTLS && pc.TLSCfg == nil {
		fmt.Fprintf(os.Stderr, "Producer missing TLS config\n")
		*errCount++
	}
	if pc.ProdFlushFreq < 0 {
		fmt.Fprintf(os.Stderr, "Producer ProdFlushFreq less than zero\n")
		*errCount++
	}
	if pc.ProdRetryMax < 0 {
		fmt.Fprintf(os.Stderr, "Producer ProdRetryMax less than zero\n")
		*errCount++
	}
	if pc.ProdRetryFreq < 0 {
		fmt.Fprintf(os.Stderr, "Producer ProdRetryFreq less than zero\n")
		*errCount++
	}
	if pc.MetaRetryMax < 0 {
		fmt.Fprintf(os.Stderr, "Producer MetaRetryMax less than zero\n")
		*errCount++
	}
	if pc.MetaRetryFreq < 0 {
		fmt.Fprintf(os.Stderr, "Producer MetaRetryFreq less than zero\n")
		*errCount++
	}
}

func checkRotationConfig(rc RotationConfiguration, errCount *int) {
	if rc.MaxSize < 0 {
		fmt.Fprintf(os.Stderr, "Rotation MaxSize less than zero\n")
		*errCount++
	}
	if rc.MaxAge < 0 {
		fmt.Fprintf(os.Stderr, "Rotation MaxAge less than zero\n")
		*errCount++
	}
	if rc.MaxBackups < 0 {
		fmt.Fprintf(os.Stderr, "Rotation MaxBackups less than zero\n")
		*errCount++
	}
}

func checkLoggerTypes(lc LoggerConfiguration, errCount *int) {
	switch lc.LogPackage {
	case ZapType:
	case LogrusType:
	case "":
	default:
		fmt.Fprintf(os.Stderr, "Invalid LogPackage type: %s\n", lc.LogPackage)
		*errCount++
	}

	switch lc.LogLevel {
	case DebugType:
	case InfoType:
	case WarnType:
	case ErrorType:
	case FatalType:
	case PanicType:
	case "":
	default:
		fmt.Fprintf(os.Stderr, "Invalid LogLevel type: %s\n", lc.LogLevel)
		*errCount++
	}

	switch lc.ConsoleFormat {
	case JSONFormat:
	case TextFormat:
	case "":
	case CEFormat:
		fallthrough
	default:
		fmt.Fprintf(os.Stderr, "Invalid ConsoleFormat type: %s\n",
			lc.ConsoleFormat)
		*errCount++
	}

	switch lc.ConsoleWriter {
	case Stdout:
	case Stderr:
	case "":
	default:
		fmt.Fprintf(os.Stderr, "Invalid ConsoleWriter type: %s\n",
			lc.ConsoleWriter)
		*errCount++
	}

	switch lc.KafkaFormat {
	case JSONFormat:
	case TextFormat:
	case CEFormat:
	case "":
	default:
		fmt.Fprintf(os.Stderr, "Invalid KafkaFormat type: %s\n", lc.KafkaFormat)
		*errCount++
	}

	switch lc.FileFormat {
	case JSONFormat:
	case TextFormat:
	case "":
	case CEFormat:
		fallthrough
	default:
		fmt.Fprintf(os.Stderr, "Invalid FileFormat type: %s\n", lc.FileFormat)
		*errCount++
	}
}

func checkCETypes(cc CloudEventsConfiguration, errCount *int) {
	switch cc.SetID {
	case CEHMAC:
	case CEUUID:
	case CEIncrID:
	case CEFuncID:
	case "":
	default:
		fmt.Fprintf(os.Stderr, "Invalid SetID type: %s\n", cc.SetID)
		*errCount++
	}
}

func checkProducerTypes(pc ProducerConfiguration, errCount *int) {
	switch pc.Partition {
	case RandomPartition:
	case HashPartition:
	case RoundRobinPartition:
	case "":
	default:
		fmt.Fprintf(os.Stderr, "Invalid Partition type: %s\n", pc.Partition)
		*errCount++
	}

	switch pc.Key {
	case LevelKey:
	case TimeSecondKey:
	case TimeNanoSecondKey:
	case FixedKey:
	case ExtractedKey:
	case FunctionKey:
	case "":
	default:
		fmt.Fprintf(os.Stderr, "Invalid Key type: %s\n", pc.Key)
		*errCount++
	}

	switch pc.Compression {
	case CompressionNone:
	case CompressionGZIP:
	case CompressionSnappy:
	case CompressionLZ4:
	case CompressionZSTD:
	case "":
	default:
		fmt.Fprintf(os.Stderr, "Invalid Compression type: %s\n", pc.Compression)
		*errCount++
	}

	switch pc.AckWait {
	case WaitForNone:
	case WaitForLocal:
	case WaitForAll:
	case "":
	default:
		fmt.Fprintf(os.Stderr, "Invalid AckWait type: %s\n", pc.AckWait)
		*errCount++
	}
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
