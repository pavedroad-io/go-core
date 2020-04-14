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
	cluster "github.com/bsm/sarama-cluster"
	"gopkg.in/yaml.v2"
)

type TestCases struct {
	Name string
	Desc string
}

var update = flag.Bool("update", false, "update golden files")

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
		fmt.Printf("Failed to exec docker-compose %+v: %s\n", myargs, err.Error())
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

var debug = flag.Bool("d", false, "Enable debug")
var rewrite = flag.Bool("r", false, "Rewrite config")

func TestMain(m *testing.M) {
	var (
		err     error
		code    int
		pubsub  bool
		subtest string
	)

	flag.Parse()
	listval := flag.Lookup("test.list").Value.String()
	runval := flag.Lookup("test.run").Value.String()
	runsplit := strings.Split(runval, "/")
	runtest := runsplit[0]
	if len(runsplit) > 1 {
		subtest = runsplit[1]
	}

	if *debug {
		fmt.Printf("=== INFO  Debug enabled\n")
		fmt.Printf("--- runtest = <%+v>\n", runtest)
		fmt.Printf("--- subtest = <%+v>\n", subtest)
	}

	if *rewrite {
		fmt.Printf("=== INFO  Rewriting config files\n")
	}

	if testing.Short() {
		// Skip Pubsub tests in short mode
		pubsub = false
	} else if runtest == "" && subtest == "" {
		pubsub = true
	} else {
		pubsub = regexp.MustCompile(runtest).MatchString("TestPubsub") ||
			subtest != "" && regexp.MustCompile(subtest).MatchString("Pubsub")
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

func TestConsole(t *testing.T) {
	var testcases = []TestCases{
		{"LogrusConsoleDefault", "logrus logger to console with default config"},
		{"ZapConsoleDefault", "zap logger to console with default config"},
	}

	for _, tc := range testcases {
		t.Run(tc.Name, func(t *testing.T) {
			var containsUnsortedJSON bool
			stdout := filepath.Join("testdata", tc.Name+".out")
			fstdout, err := os.Create(stdout)
			defer fstdout.Close()
			if err != nil {
				t.Fatalf("Failed to create file %s: %s\n", stdout, err.Error())
			}
			saveout := os.Stdout
			os.Stdout = fstdout

			cfg, err := readConfiguration(t, tc.Name)
			if err != nil {
				t.FailNow()
			}

			err = executeTests(t, cfg)
			if err != nil {
				t.FailNow()
			}

			os.Stdout = saveout

			var actual []byte
			if containsUnsortedJSON {
				actual, err = normalizeJSONFile(t, stdout)
				if err != nil {
					t.FailNow()
				}
			} else {
				actual, err = ioutil.ReadFile(stdout)
				if err != nil {
					t.Fatalf("Failed to read file %s: %s\n", stdout,
						err.Error())
				}
			}

			golden := filepath.Join("testdata", tc.Name+".golden")
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
		})
	}
}

func TestLogfile(t *testing.T) {
	var testcases = []TestCases{
		{"LogrusLogfileDefault", "logrus logger to file with default config"},
		{"ZapLogfileDefault", "zap logger to file with default config"},
	}

	for _, tc := range testcases {
		t.Run(tc.Name, func(t *testing.T) {
			cfg, err := readConfiguration(t, tc.Name)
			if err != nil {
				t.FailNow()
			}

			flogfile, err := os.Create(cfg.FileLocation)
			defer flogfile.Close()
			if err != nil {
				t.Fatalf("Failed to create file %s: %s\n",
					cfg.FileLocation, err.Error())
			}

			err = executeTests(t, cfg)
			if err != nil {
				t.FailNow()
			}

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

			golden := filepath.Join("testdata", tc.Name+".golden")
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
		})
	}
}

func TestPubsub(t *testing.T) {
	var testcases = []TestCases{
		{"LogrusPubsubDefault", "logrus logger to kafka with default config"},
		{"ZapPubsubDefault", "zap logger to kafka with default config"},
		{"LogrusPubsubTopic", "logrus logger to kafka with test topic"},
		{"ZapPubsubTopic", "zap logger to kafka with test topic"},
		{"LogrusPubsubIncrID", "logrus logger to kafka with incremental ID"},
		{"ZapPubsubIncrID", "zap logger to kafka with incremental ID"},
	}

	var (
		brokers = []string{"localhost:9092"}
		group   = "testgroup"
		topics  = []string{"logs", "test"}
		config  = cluster.NewConfig()
	)

	if testing.Short() {
		// Skip Pubsub tests in short mode
		t.SkipNow()
	}

	time.Sleep(5 * time.Second)
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	consumer, err := cluster.NewConsumer(brokers, group, topics, config)
	if err != nil {
		t.Errorf("Failed to initialize consumer: %s\n", err.Error())
	}
	defer consumer.Close()
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	for _, tc := range testcases {
		t.Run(tc.Name, func(t *testing.T) {
			topictest := regexp.MustCompile("Topic").MatchString(tc.Name)
			cfg, err := readConfiguration(t, tc.Name)
			if err != nil {
				t.FailNow()
			}

			if topictest {
				err = executeTopicTests(t, cfg, "test")
			} else {
				err = executeTests(t, cfg)
			}
			if err != nil {
				t.FailNow()
			}

			var actual []byte
			var message string
			time.Sleep(2 * time.Second)
			done := time.After(2 * time.Second)
		readpubsub:
			for {
				select {
				case msg := <-consumer.Messages():
					consumer.MarkOffset(msg, "kafka-test")
					message = fmt.Sprintf("T:%s P:%d K:%s V:%s\n",
						msg.Topic, msg.Partition, msg.Key, msg.Value)
					actual = append(actual, message...)
					if *debug {
						t.Log(message)
					}
				case err := <-consumer.Errors():
					t.Logf("Consumer message error: %s\n", err.Error())
				case <-interrupt:
					t.Logf("Consumer caught interrupt signal\n")
					break readpubsub
				case <-done:
					break readpubsub
				}
			}

			pub := filepath.Join("testdata", tc.Name+".pub")
			ioutil.WriteFile(pub, actual, 0644)

			golden := filepath.Join("testdata", tc.Name+".golden")
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
		})
	}
}
