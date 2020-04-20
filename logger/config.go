package logger

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"os/user"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

// configType provides configuration type
type configType string

// Supported configuration types
const (
	EnvConfig  configType = "env"
	FileConfig configType = "file"
	BothConfig configType = "both"
)

// Supported auto init/config environment names
const (
	LogAutoInitEnvName = "PRLOG_AUTOINIT"
	ConfigTypeEnvName  = "PRLOG_CFGTYPE"
	ConfigFileEnvName  = "PRLOG_CFGFILE"
)

// Supported environment name prefixes
const (
	LogEnvPrefix         = "PRLOG"
	KafkaEnvPrefix       = "PRKAFKA"
	CloudEventsEnvPrefix = "PRCE"
	RotationEnvPrefix    = "PRROT"
)

// Default config file name without extension
// Config file name exported on SIGUSR1
const (
	ConfigFileName       = "pr_log_config"
	ExportConfigFileName = "pr_export_config.yaml"
)

// Supported error messages
const (
	errInvalid     = "Invalid configuration type"
	errLogger      = "Could not create logger configuration"
	errKafka       = "Could not create kafka configuration"
	errCloudevents = "Could not create cloudevents configuration"
	errRotation    = "Could not create rotation configuration"
)

// logger global for go log pkg emulation
var logger Logger

var ErrFatal = errors.New("fatal")
var ErrNonFatal = errors.New("nonfatal")

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
	EnableDebug:       false,
}

var defaultProducerConfiguration = ProducerConfiguration{
	Brokers:       []string{"localhost:9092"},
	Topic:         "logs",
	Partition:     RandomPartition,
	Key:           FixedKey,
	KeyName:       "username",
	Compression:   CompressionSnappy,
	AckWait:       WaitForLocal,
	ProdFlushFreq: 500 * time.Millisecond,
	ProdRetryMax:  10,
	ProdRetryFreq: 100 * time.Millisecond,
	MetaRetryMax:  10,
	MetaRetryFreq: 2000 * time.Millisecond,
	EnableTLS:     false,
	EnableDebug:   false,
}

var defaultCloudEventsConfiguration = CloudEventsConfiguration{
	SetID:           CEHMAC,
	HMACKey:         "pavedroad",
	Source:          "http://github.com/pavedroad-io/go-core/logger",
	SpecVersion:     "1.0",
	Type:            "io.pavedroad.cloudevents.log",
	SetSubjectLevel: true,
}

var defaultRotationConfiguration = RotationConfiguration{
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

// DefaultLoggerCfg returns default log configuration
func DefaultCompleteCfg() *LoggerConfiguration {
	config := defaultLoggerConfiguration
	config.CloudEventsCfg = defaultCloudEventsConfiguration
	config.KafkaProducerCfg = defaultProducerConfiguration
	config.RotationCfg = defaultRotationConfiguration
	return &config
}

// init called on package import to configure and initialize default logger
func init() {
	var err error

	// set PRLOG_AUTOINIT=true to initialize logger with default configuration
	autoInit := os.Getenv(LogAutoInitEnvName)
	if autoInit != "true" {
		return
	}

	// set PRLOG_CFGTYPE as needed to specify how to override logger defaults
	cfgType := configType(os.Getenv(ConfigTypeEnvName))
	if cfgType == "" {
		// default to override configuration defaults via environment
		cfgType = EnvConfig
	}

	// set PRLOG_CFGFILE to override default config file name
	cfgFile := os.Getenv(ConfigFileEnvName)
	if cfgFile == "" {
		cfgFile = ConfigFileName
	}

	config, err := GetLoggerConfiguration(cfgType, cfgFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		if errors.Is(err, ErrFatal) {
			os.Exit(1)
		}
	}

	// initialize the logger with the customized configuration
	if logger, err = NewLogger(config); err != nil {
		fmt.Fprintf(os.Stderr, "Could not instantiate %s logger package: %s\n",
			config.LogPackage, err.Error())
		os.Exit(1)
	}
}

// GetLoggerConfiguration generates config from defaults/config-file/environment
func GetLoggerConfiguration(cfgType configType,
	cfgFileName string) (LoggerConfiguration, error) {
	var cfg LoggerConfiguration
	errSetting := ErrNonFatal

	switch cfgType {
	case EnvConfig:
	case FileConfig:
	case BothConfig:
	default:
		return cfg, fmt.Errorf("%s: %s %w\n", errInvalid, cfgType, ErrFatal)
	}

	user, err := user.Current()
	if err == nil {
		defaultProducerConfiguration.KeyName = user.Username
	}

	config := new(LoggerConfiguration)
	// read config file and/or environment to override defaults
	// single config file covers basic log config and all sub configs
	// only gets environment overrides for the basic log config
	err = FillConfiguration(DefaultCompleteCfg(), config, cfgType, cfgFileName,
		LogEnvPrefix)
	if err != nil {
		return cfg, fmt.Errorf("%s: %s %w\n", errLogger, err.Error(), ErrFatal)
	}

	// get environment overrides for the kafka sub config
	kafkaConfig := new(ProducerConfiguration)
	err = FillConfiguration(DefaultProducerCfg(), kafkaConfig, EnvConfig, "",
		KafkaEnvPrefix)
	if err == nil {
		config.KafkaProducerCfg = *kafkaConfig
	} else {
		if config.EnableKafka {
			errSetting = ErrFatal
		}
		return cfg, fmt.Errorf("%s: %s %w\n", errKafka, err.Error(), errSetting)
	}

	// get environment overrides for the cloudevents sub config
	ceConfig := new(CloudEventsConfiguration)
	err = FillConfiguration(DefaultCloudEventsCfg(), ceConfig, EnvConfig, "",
		CloudEventsEnvPrefix)
	if err == nil {
		config.CloudEventsCfg = *ceConfig
	} else {
		if config.EnableCloudEvents {
			errSetting = ErrFatal
		}
		return cfg, fmt.Errorf("%s: %s %w\n", errCloudevents, err.Error(),
			errSetting)
	}

	// get environment overrides for the rotation sub config
	rotConfig := new(RotationConfiguration)
	err = FillConfiguration(DefaultRotationCfg(), rotConfig, EnvConfig, "",
		RotationEnvPrefix)
	if err == nil {
		config.RotationCfg = *rotConfig
	} else {
		if config.EnableRotation {
			errSetting = ErrFatal
		}
		return cfg, fmt.Errorf("%s: %s %w\n", errRotation, err.Error(),
			errSetting)
	}
	return *config, nil
}

// FillConfiguration fills config from defaults, config file and environment
func FillConfiguration(defaultCfg interface{}, config interface{},
	cfgType configType, filename string, prefix string) error {

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
	if cfgType == EnvConfig || cfgType == BothConfig {
		v.SetEnvPrefix(prefix)
		v.AutomaticEnv()
	}

	if cfgType == FileConfig || cfgType == BothConfig {
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

func ExportConfiguration(file string, config LoggerConfiguration) error {

	ybytes, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("Failed to marshal config %s\n", err.Error())
	}
	if file != "" {
		err = ioutil.WriteFile(file, ybytes, 0644)
		if err != nil {
			return fmt.Errorf("Failed to export config %s\n", err.Error())
		}
	}
	if config.EnableDebug {
		os.Stderr.Write(ybytes)
	}
	return nil
}

var globalLoggerConfiguration LoggerConfiguration

func signalCatcher() {
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGUSR1)
	<-ch
	ExportConfiguration(ExportConfigFileName, globalLoggerConfiguration)
	go signalCatcher()
}

func checkConfig(config LoggerConfiguration) error {
	var errCount int

	globalLoggerConfiguration = config
	go signalCatcher()
	if config.EnableDebug {
		ExportConfiguration("", config)
	}

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
	if logger == nil {
		fmt.Fprintf(os.Stderr, "Logger not initialized\n")
		return
	}
	logger.Info(args...)
}

// Printf emulates function from go log pkg
func Printf(format string, args ...interface{}) {
	if logger == nil {
		fmt.Fprintf(os.Stderr, "Logger not initialized\n")
		return
	}
	logger.Infof(format, args...)
}

// Println emulates function from go log pkg
func Println(args ...interface{}) {
	if logger == nil {
		fmt.Fprintf(os.Stderr, "Logger not initialized\n")
		return
	}
	logger.Info(strings.TrimRight(fmt.Sprintln(args...), "\n"))
}

// Debug emulates function from go log pkg
func Debug(args ...interface{}) {
	if logger == nil {
		fmt.Fprintf(os.Stderr, "Logger not initialized\n")
		return
	}
	logger.Debug(args...)
}

// Debugf emulates function from go log pkg
func Debugf(format string, args ...interface{}) {
	if logger == nil {
		fmt.Fprintf(os.Stderr, "Logger not initialized\n")
		return
	}
	logger.Debugf(format, args...)
}

// Debugln emulates function from go log pkg
func Debugln(args ...interface{}) {
	if logger == nil {
		fmt.Fprintf(os.Stderr, "Logger not initialized\n")
		return
	}
	logger.Debug(strings.TrimRight(fmt.Sprintln(args...), "\n"))
}

// Info emulates function from go log pkg
func Info(args ...interface{}) {
	if logger == nil {
		fmt.Fprintf(os.Stderr, "Logger not initialized\n")
		return
	}
	logger.Info(args...)
}

// Infof emulates function from go log pkg
func Infof(format string, args ...interface{}) {
	if logger == nil {
		fmt.Fprintf(os.Stderr, "Logger not initialized\n")
		return
	}
	logger.Infof(format, args...)
}

// Infoln emulates function from go log pkg
func Infoln(args ...interface{}) {
	if logger == nil {
		fmt.Fprintf(os.Stderr, "Logger not initialized\n")
		return
	}
	logger.Info(strings.TrimRight(fmt.Sprintln(args...), "\n"))
}

// Warn emulates function from go log pkg
func Warn(args ...interface{}) {
	if logger == nil {
		fmt.Fprintf(os.Stderr, "Logger not initialized\n")
		return
	}
	logger.Warn(args...)
}

// Warnf emulates function from go log pkg
func Warnf(format string, args ...interface{}) {
	if logger == nil {
		fmt.Fprintf(os.Stderr, "Logger not initialized\n")
		return
	}
	logger.Warnf(format, args...)
}

// Warnln emulates function from go log pkg
func Warnln(args ...interface{}) {
	if logger == nil {
		fmt.Fprintf(os.Stderr, "Logger not initialized\n")
		return
	}
	logger.Warn(strings.TrimRight(fmt.Sprintln(args...), "\n"))
}

// Error emulates function from go log pkg
func Error(args ...interface{}) {
	if logger == nil {
		fmt.Fprintf(os.Stderr, "Logger not initialized\n")
		return
	}
	logger.Error(args...)
}

// Errorf emulates function from go log pkg
func Errorf(format string, args ...interface{}) {
	if logger == nil {
		fmt.Fprintf(os.Stderr, "Logger not initialized\n")
		return
	}
	logger.Errorf(format, args...)
}

// Errorln emulates function from go log pkg
func Errorln(args ...interface{}) {
	if logger == nil {
		fmt.Fprintf(os.Stderr, "Logger not initialized\n")
		return
	}
	logger.Error(strings.TrimRight(fmt.Sprintln(args...), "\n"))
}

// Fatal emulates function from go log pkg
func Fatal(args ...interface{}) {
	if logger == nil {
		fmt.Fprintf(os.Stderr, "Logger not initialized\n")
		return
	}
	logger.Fatal(args...)
}

// Fatalf emulates function from go log pkg
func Fatalf(format string, args ...interface{}) {
	if logger == nil {
		fmt.Fprintf(os.Stderr, "Logger not initialized\n")
		return
	}
	logger.Fatalf(format, args...)
}

// Fatalln emulates function from go log pkg
func Fatalln(args ...interface{}) {
	if logger == nil {
		fmt.Fprintf(os.Stderr, "Logger not initialized\n")
		return
	}
	logger.Fatal(strings.TrimRight(fmt.Sprintln(args...), "\n"))
}

// Panic emulates function from go log pkg
func Panic(args ...interface{}) {
	if logger == nil {
		fmt.Fprintf(os.Stderr, "Logger not initialized\n")
		return
	}
	logger.Panic(args...)
}

// Panicf emulates function from go log pkg
func Panicf(format string, args ...interface{}) {
	if logger == nil {
		fmt.Fprintf(os.Stderr, "Logger not initialized\n")
		return
	}
	logger.Fatalf(format, args...)
}

// Panicln emulates function from go log pkg
func Panicln(args ...interface{}) {
	if logger == nil {
		fmt.Fprintf(os.Stderr, "Logger not initialized\n")
		return
	}
	logger.Panic(strings.TrimRight(fmt.Sprintln(args...), "\n"))
}
