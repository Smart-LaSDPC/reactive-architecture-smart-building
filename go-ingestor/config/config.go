package config

import (
	"fmt"
	"io/ioutil"
	"time"
	"os"

	"gopkg.in/yaml.v2"
)

type AppConfig struct {
	Kafka struct {
		BrokerAddress    string `yaml:"brokerAddress"`
		Version          string `yaml:"version"`
		ConsumerGroupID  string `yaml:"consumerGroupID"`
		Topic            string `yaml:"topic"`
		AssignorStrategy string `yaml:"assignorStrategy"`
		OffsetOldest     bool   `yaml:"offsetOldest"`
		Verbose          bool   `yaml:"verbose"`
	} `yaml:"kafka"`
	DB struct {
		Host      	 	  string `yaml:"host"`
		Port      	 	  string `yaml:"port"`
		User      	 	  string `yaml:"user"`
		Password  	 	  string `yaml:"password"`
		SslMode   	 	  string `yaml:"sslMode"`
		DbName    	 	  string `yaml:"dbName"`
		TableName 	 	  string `yaml:"tableName"`
		Conns  	 		  int32  `yaml:"conns"`
		QueryTimeout 	  time.Duration `yaml:"queryTimeout"`
		BatchSize    	  int `yaml:"batchSize"`
		BatchFlushTimeout time.Duration `yaml:"batchFlushTimeout"`
	} `yaml:"db"`
}

func GetAppConfig() (*AppConfig, error) {
	var err error

	c := &AppConfig{}

	configFile := ReadEnvVariable("CONFIG_FILE", "config/config.yaml")

	c, err = c.setFromYamlFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("cannot read YAML file %s: %s", configFile, err)
	}

	return c, nil
}

func (c *AppConfig) setFromYamlFile(filename string) (*AppConfig, error) {
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	if err = yaml.Unmarshal(buf, c); err != nil {
		return nil, err
	}

	return c, nil
}

func ReadEnvVariable(env, fallback string) string {
	v, ok := os.LookupEnv(env)
	if !ok {
		return fallback
	}
	return v
}
