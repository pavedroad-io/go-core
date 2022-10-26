package logger

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/Shopify/sarama"
	"gopkg.in/yaml.v2"
)

var (
	update   = flag.Bool("update", false, "update golden files")
	debug    = flag.Bool("d", false, "Enable debug")
	rewrite  = flag.Bool("r", false, "Rewrite config")
	explicit = flag.String("t", "", "Explicit test")
)

var (
	testenv     bool
	testinit    bool
	testshort   bool
	testverbose bool
)

func TestMain(m *testing.M) {
	var (
		err     error
		code    int
		pubsub  bool
		subtest string
	)
	flag.Parse()

	testEnv := os.Getenv("PRTEST_ENV")
	if testEnv == "true" {
		testenv = true
	}
	testInit := os.Getenv("PRTEST_INIT")
	if testInit == "true" {
		testinit = true
	}
	if testing.Short() {
		testshort = true
	}
	if testing.Verbose() {
		testverbose = true
	}

	listval := flag.Lookup("test.list").Value.String()
	runval := flag.Lookup("test.run").Value.String()
	runsplit := strings.Split(runval, "/")
	runtest := runsplit[0]
	if len(runsplit) > 1 {
		subtest = runsplit[1]
	}

	if testshort {
		fmt.Printf("=== INFO  Short enabled\n")
	}
	if testverbose {
		fmt.Printf("=== INFO  Verbose enabled\n")
	}
	if *debug {
		fmt.Printf("=== INFO  Debug enabled\n")
		fmt.Printf("--- explicit = <%+v>\n", *explicit)
		fmt.Printf("--- runtest = <%+v>\n", runtest)
		fmt.Printf("--- subtest = <%+v>\n", subtest)
	}

	if *rewrite {
		fmt.Printf("=== INFO  Rewriting config files\n")
	}

	if testenv || testinit || testshort {
		// Server started externally for env/init tests
		// Skip Pubsub tests in short mode
		pubsub = false
	} else if runtest == "" && subtest == "" {
		pubsub = true
	} else {
		pubsub = regexp.MustCompile(tPub).MatchString(runtest) ||
			subtest != "" && regexp.MustCompile(tPub).MatchString(subtest)
	}

	if pubsub && listval == "" {
		fmt.Printf("=== START Pubsub server\n")
		err = pubsubStartup()
	}

	if err == nil {
		code = m.Run()
	} else {
		code = 1
	}

	if pubsub && listval == "" {
		fmt.Printf("=== STOP  Pubsub server\n")
		pubsubShutdown()
	}

	os.Exit(code)
}

func readConfiguration(t *testing.T, testname string) (
	LoggerConfiguration, error) {

	var cfg LoggerConfiguration
	input := filepath.Join("testdata", testname+".yaml")
	yamlbytes, err := ioutil.ReadFile(input)
	if err != nil {
		t.Errorf("Failed to read %s config: %s\n", input, err.Error())
		return cfg, err
	}

	err = yaml.Unmarshal(yamlbytes, &cfg)
	if err != nil {
		t.Errorf("Failed to unmarshal %s config %s\n", input, err.Error())
		return cfg, err
	}
	if *rewrite {
		err = writeConfiguration(t, input, cfg)
		return cfg, err
	}
	return cfg, nil
}

func writeConfiguration(t *testing.T, file string,
	cfg LoggerConfiguration) error {

	ybytes, err := yaml.Marshal(cfg)
	if err != nil {
		t.Errorf("Failed to marshal %s config %s\n", file, err.Error())
		return err
	}
	ioutil.WriteFile(file, ybytes, 0644)
	return nil
}

func executeTests(t *testing.T, cfg LoggerConfiguration) error {
	log, err := NewLogger(cfg)

	if err != nil {
		t.Errorf("Failed to instantiate %s logger: %s",
			cfg.LogPackage, err.Error())
		return err
	}

	log.Debugf("Debugf using %s", "Debugf (should not appear)")
	log.Infof("Infof using %s", cfg.LogPackage)
	log.Warnf("Warnf using %s", cfg.LogPackage)
	log.Errorf("Errorf using %s", cfg.LogPackage)
	log.Print("Print using", cfg.LogPackage)
	log.Printf("Printf using %s", cfg.LogPackage)
	log.Println("Println using", cfg.LogPackage)
	return nil
}

func executeInitTests(t *testing.T, cfg LoggerConfiguration) error {
	Debugf("Debugf using %s", "Debugf (should not appear)")
	Infof("Infof using %s", cfg.LogPackage)
	Warnf("Warnf using %s", cfg.LogPackage)
	Errorf("Errorf using %s", cfg.LogPackage)
	Print("Print using", cfg.LogPackage)
	Printf("Printf using %s", cfg.LogPackage)
	Println("Println using", cfg.LogPackage)
	return nil
}

func executeTopicTests(t *testing.T, cfg LoggerConfiguration,
	topic string) error {

	log, err := NewLogger(cfg)
	if err != nil {
		t.Errorf("Failed to instantiate %s logger: %s",
			cfg.LogPackage, err.Error())
		return err
	}

	topicfield := LogFields{TopicKey: topic}
	log.WithFields(topicfield).Infof("Infof using %s", cfg.LogPackage)
	log.WithFields(topicfield).Warnf("Warnf using %s", cfg.LogPackage)
	log.WithFields(topicfield).Errorf("Errorf using %s", cfg.LogPackage)
	log.WithFields(topicfield).Print("Print using", cfg.LogPackage)
	log.WithFields(topicfield).Printf("Printf using %s", cfg.LogPackage)
	log.WithFields(topicfield).Println("Println using", cfg.LogPackage)
	return nil
}

func normalizeJSONFile(t *testing.T, filename string) ([]byte, error) {
	var jsonbytes []byte

	jsonlog, err := os.Open(filename)
	defer jsonlog.Close()
	if err != nil {
		t.Errorf("Failed to open %s: %s", filename, err.Error())
		return nil, err
	}

	scanner := bufio.NewScanner(jsonlog)
	for scanner.Scan() {
		normbytes, err := normalizeJSONLine(t, scanner.Text())
		if err != nil {
			t.Errorf("Failed to normalize %s: %s\n",
				scanner.Text(), err.Error())
			return nil, err
		}
		jsonbytes = append(jsonbytes, normbytes...)
	}
	if err := scanner.Err(); err != nil {
		t.Errorf("Failed to scan %s: %s", filename, err.Error())
		return nil, err
	}
	return jsonbytes, nil
}

func normalizeJSONLine(t *testing.T, line string) ([]byte, error) {
	var jsonmap map[string]interface{}
	var prejson, jsontext string

	index := strings.IndexByte(line, '{')
	if index == -1 {
		t.Errorf("Failed to find JSON in line <%s>\n", line)
		return nil, errors.New("JSON scan failure")
	} else if index == 0 {
		jsontext = line
	} else {
		prejson = line[:index]
		jsontext = line[index:]
	}
	err := json.Unmarshal([]byte(jsontext), &jsonmap)
	if err != nil {
		t.Errorf("Failed to unmarshal %+v err: %s\n", jsontext, err.Error())
		return nil, err
	}
	jsonbytes, err := json.Marshal(jsonmap)
	if err != nil {
		t.Errorf("Failed to marshal %+v err: %s\n", jsonmap, err.Error())
		return nil, err
	}
	jsonbytes = append(jsonbytes, "\n"...)
	return append([]byte(prejson), jsonbytes...), nil
}

func dockerCompose(file string, args ...string) error {
	var myargs []string
	if file != "" {
		myargs = append(myargs, "-f", file)
	}
	myargs = append(myargs, args...)

	cmd := exec.Command("docker-compose", myargs...)
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Failed to exec docker-compose %+v: %s\n", myargs,
			err.Error())
		return err
	}
	return nil
}

func pubsubStartup() error {
	err := dockerCompose("testdata/docker-compose.yaml", "up", "-d")
	if err != nil {
		fmt.Printf("Failed to startup kafka server\n")
	}
	return err
}

func pubsubShutdown() error {
	err := dockerCompose("testdata/docker-compose.yaml", "down")
	if err != nil {
		fmt.Printf("Failed to shutdown kafka server\n")
	}
	return err
}

func setupConsole(t *testing.T, name string, pkg string,
	cfg LoggerConfiguration) *os.File {

	if testinit {
		return nil
	}

	fname := filepath.Join("testdata", name+".out")
	file, err := os.Create(fname)
	if err != nil {
		t.Fatalf("Failed to create file %s: %s\n", fname, err.Error())
	}

	var output *os.File
	if cfg.ConsoleWriter == Stderr {
		output = os.Stderr
		os.Stderr = file
	} else {
		output = os.Stdout
		os.Stdout = file
	}
	return output
}

func checkConsole(t *testing.T, name string, pkg string,
	cfg LoggerConfiguration, output *os.File) {
	var containsUnsortedJSON bool
	var err error

	if output != nil {
		if cfg.ConsoleWriter == Stderr {
			os.Stderr.Close()
			os.Stderr = output
		} else {
			os.Stdout.Close()
			os.Stdout = output
		}
	}

	var actual []byte
	fname := filepath.Join("testdata", name+".out")
	if containsUnsortedJSON {
		actual, err = normalizeJSONFile(t, fname)
		if err != nil {
			t.FailNow()
		}
	} else {
		actual, err = ioutil.ReadFile(fname)
		if err != nil {
			t.Fatalf("Failed to read file %s: %s\n", fname,
				err.Error())
		}
	}

	golden := filepath.Join("testdata", name+".golden")
	if *update {
		t.Logf("Updating %s\n", golden)
		ioutil.WriteFile(golden, actual, 0644)
	}

	expected, err := ioutil.ReadFile(golden)
	if err != nil {
		t.Fatalf("Failed to read file %s: %s\n", golden, err.Error())
	}

	if !bytes.Equal(actual, expected) {
		t.FailNow()
	}
}

func setupLogfile(t *testing.T, name string, pkg string,
	cfg LoggerConfiguration) *os.File {

	flogfile, err := os.Create(cfg.FileLocation)
	if err != nil {
		t.Fatalf("Failed to create file %s: %s\n",
			cfg.FileLocation, err.Error())
	}
	return flogfile
}

func checkLogfile(t *testing.T, name string, pkg string,
	cfg LoggerConfiguration, output *os.File) {
	var err error

	output.Close()
	var actual []byte
	if cfg.LogPackage == ZapType {
		actual, err = normalizeJSONFile(t, cfg.FileLocation)
		if err != nil {
			t.FailNow()
		}
	} else {
		actual, err = ioutil.ReadFile(cfg.FileLocation)
		if err != nil {
			t.Fatalf("Failed to read file %s: %s\n",
				cfg.FileLocation, err.Error())
		}
	}

	golden := filepath.Join("testdata", name+".golden")
	if *update {
		t.Logf("Updating %s\n", golden)
		ioutil.WriteFile(golden, actual, 0644)
	}

	expected, err := ioutil.ReadFile(golden)
	if err != nil {
		t.Fatalf("Failed to read file %s: %s\n", golden, err.Error())
	}

	if !bytes.Equal(actual, expected) {
		t.FailNow()
	}
}

func setupPubsub(t *testing.T, name string, pkg string,
	cfg LoggerConfiguration) {

	if testing.Short() {
		// Skip Pubsub tests in short mode
		t.SkipNow()
	}
	time.Sleep(5 * time.Second)
}

func checkPubsub(t *testing.T, name string, pkg string, cfg LoggerConfiguration,
	topic string, offset int64, count int64) {
	var actual []byte
	var message string

	var (
		brokers = []string{"localhost:9092"}
		config  = sarama.NewConfig()
	)
	master, err := sarama.NewConsumer(brokers, config)
	if err != nil {
		t.Errorf("Failed to initialize consumer: %s\n", err.Error())
	}
	consumer, err := master.ConsumePartition(topic, 0, offset)
	if err != nil {
		t.Errorf("Failed to get partition consumer: %s\n", err.Error())
	}
	defer consumer.Close()
	defer master.Close()

	time.Sleep(2 * time.Second)
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	done := time.After(2 * time.Second)
	var msgCount int64 = 0

readpubsub:
	for {
		select {
		case msg := <-consumer.Messages():
			message = fmt.Sprintf("T:%s P:%d K:%s V:%s\n",
				msg.Topic, msg.Partition, msg.Key, msg.Value)
			actual = append(actual, message...)
			if *debug {
				t.Log(message)
			}
			msgCount++
			if msgCount >= count {
				break readpubsub
			}
		case err := <-consumer.Errors():
			t.Logf("Consumer message error: %s\n", err.Error())
		case <-interrupt:
			t.Logf("Consumer caught interrupt signal\n")
			break readpubsub
		case <-done:
			t.Logf("Consumer caught timeout signal\n")
			break readpubsub
		}
	}

	pub := filepath.Join("testdata", name+".pub")
	ioutil.WriteFile(pub, actual, 0644)

	golden := filepath.Join("testdata", name+".golden")
	if *update {
		t.Logf("Updating %s\n", golden)
		ioutil.WriteFile(golden, actual, 0644)
	}

	expected, err := ioutil.ReadFile(golden)
	if err != nil {
		t.Fatalf("Failed to read file %s: %s\n", golden, err.Error())
	}

	if !bytes.Equal(actual, expected) {
		t.FailNow()
	}
}

func getConfiguration(t *testing.T, testname string,
	prefix string) LoggerConfiguration {
	var cfg LoggerConfiguration
	var err error

	if testenv {
		cfg, err = GetLoggerConfiguration(EnvConfig, "")
	} else if testinit {
		cfg = globalLoggerConfiguration
	} else {
		cfg, err = readConfiguration(t, testname)
	}
	if err != nil {
		t.FailNow()
	}
	return cfg
}

func runTests(t *testing.T, name string, pkg string, cfg LoggerConfiguration,
	topic string) {
	var err error

	topicTest := regexp.MustCompile("Topic").MatchString(name)
	if testinit {
		err = executeInitTests(t, cfg)
	} else if topicTest {
		err = executeTopicTests(t, cfg, topic)
	} else {
		err = executeTests(t, cfg)
	}
	if err != nil {
		t.FailNow()
	}
}

func testHarness(t *testing.T, name string, prefix string, pkg string,
	console bool, logfile bool, pubsub bool, topic string, offset int64,
	count int64) {
	var conOutput *os.File
	var logOutput *os.File

	cfg := getConfiguration(t, name, prefix)

	if console {
		conOutput = setupConsole(t, name, pkg, cfg)
	}
	if logfile {
		logOutput = setupLogfile(t, name, pkg, cfg)
	}
	if pubsub {
		setupPubsub(t, name, pkg, cfg)
	}

	runTests(t, name, pkg, cfg, topic)

	if console {
		checkConsole(t, name, pkg, cfg, conOutput)
	}
	if logfile {
		checkLogfile(t, name, pkg, cfg, logOutput)
	}
	if pubsub {
		checkPubsub(t, name, pkg, cfg, topic, offset, count)
	}
}

const (
	tNil = ""
	tLru = "Logrus"
	tZap = "Zap"
	tCon = "Console"
	tLog = "Logfile"
	tPub = "Pubsub"
	tIni = "Init"
	tEnv = "Env"
)

type TestCases struct {
	Prefix string
	Pkg    string
	Con    string
	Log    string
	Pub    string
	Topic  string
	Offset int64
	Count  int64
	Test   string
	Desc   string
}

func runTestCases(t *testing.T, testCases []TestCases) {
	for _, tc := range testCases {
		name := fmt.Sprintf("%s%s%s%s%s%s",
			tc.Prefix, tc.Pkg, tc.Con, tc.Log, tc.Pub, tc.Test)

		if *explicit != "" && *explicit != name {
			t.SkipNow()
		}
		t.Run(name, func(t *testing.T) {
			testHarness(t, name, tc.Prefix, tc.Pkg, tc.Con == tCon,
				tc.Log == tLog, tc.Pub == tPub, tc.Topic, tc.Offset, tc.Count)
		})
	}
}

func TestConsole(t *testing.T) {
	var testCases = []TestCases{
		{tNil, tLru, tCon, tNil, tNil, tNil, 0, 0, "Default",
			"logrus logger to console with default config"},
		{tNil, tZap, tCon, tNil, tNil, tNil, 0, 0, "Default",
			"zap logger to console with default config"},
	}
	if testinit || testenv {
		t.SkipNow()
	}
	runTestCases(t, testCases)
}

func TestLogfile(t *testing.T) {
	var testCases = []TestCases{
		{tNil, tLru, tNil, tLog, tNil, tNil, 0, 0, "Default",
			"logrus logger to file with default config"},
		{tNil, tZap, tNil, tLog, tNil, tNil, 0, 0, "Default",
			"zap logger to file with default config"},
	}
	if testinit || testenv {
		t.SkipNow()
	}
	runTestCases(t, testCases)
}

func TestPubsub(t *testing.T) {
	var testCases = []TestCases{
		{tNil, tLru, tNil, tNil, tPub, "logs", 0, 6, "Default",
			"logrus logger to kafka with default config"},
		{tNil, tZap, tNil, tNil, tPub, "logs", 6, 6, "Default",
			"zap logger to kafka with default config"},
		{tNil, tLru, tNil, tNil, tPub, "test", 0, 6, "Topic",
			"logrus logger to kafka with test topic"},
		{tNil, tZap, tNil, tNil, tPub, "test", 6, 6, "Topic",
			"zap logger to kafka with test topic"},
		{tNil, tLru, tNil, tNil, tPub, "logs", 12, 6, "IncrID",
			"logrus logger to kafka with incremental ID"},
		{tNil, tZap, tNil, tNil, tPub, "logs", 18, 6, "IncrID",
			"zap logger to kafka with incremental ID"},
	}
	if testinit || testenv || testshort {
		t.SkipNow()
	}
	runTestCases(t, testCases)
}

func TestEnv(t *testing.T) {
	var testCases = []TestCases{
		{tEnv, tLru, tCon, tNil, tNil, tNil, 0, 0, "Default", tNil},
		{tEnv, tLru, tNil, tLog, tNil, tNil, 0, 0, "Default", tNil},
		{tEnv, tLru, tNil, tNil, tPub, "logs", 0, 6, "Default", tNil},
		{tEnv, tZap, tCon, tNil, tNil, tNil, 0, 0, "Default", tNil},
		{tEnv, tZap, tNil, tLog, tNil, tNil, 0, 0, "Default", tNil},
		{tEnv, tZap, tNil, tNil, tPub, "logs", 6, 6, "Default", tNil},
	}
	if !testenv {
		t.SkipNow()
	}
	runTestCases(t, testCases)
}

func TestInit(t *testing.T) {
	var testCases = []TestCases{
		{tIni, tLru, tCon, tNil, tNil, tNil, 0, 0, "Default", tNil},
		{tIni, tLru, tNil, tLog, tNil, tNil, 0, 0, "Default", tNil},
		{tIni, tLru, tNil, tNil, tPub, "logs", 12, 6, "Default", tNil},
		{tIni, tZap, tCon, tNil, tNil, tNil, 0, 0, "Default", tNil},
		{tIni, tZap, tNil, tLog, tNil, tNil, 0, 0, "Default", tNil},
		{tIni, tZap, tNil, tNil, tPub, "logs", 18, 6, "Default", tNil},
	}
	if !testinit {
		t.SkipNow()
	}
	runTestCases(t, testCases)
}
