package config

import (
	"fmt"
	"io/ioutil"
	"os"

	// "gopkg.in/validator.v2"
	"gopkg.in/yaml.v2"
)


type AppConfig struct {
	Kafka struct {
		Brokers 	     string `yaml:"brokers"`
		Version		     string `yaml:"version"`
		ConsumerGroupID  string `yaml:"ConsumerGroupID"`
		Topic            string `yaml:"topic"`
		AssignorStrategy string `yaml:"assignorStrategy"`
		OffsetOldest	 bool   `yaml:"offsetOldest"`
		Verbose			 bool   `yaml:"verbose"`
	} `yaml:"kafka"`
}

func GetAppConfig() (*AppConfig, error) {
	var err error

	c := &AppConfig{}

	configFile := ReadEnvVariable("CONFIG_FILE", "config/config.yaml")
	
	c, err = c.setFromYamlFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("cannot read YAML file %s: %s", configFile, err)
	}

	// if err := validator.Validate(c); err != nil {
	// 	return nil, fmt.Errorf("invalid configuration: %s", err)
	// }

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
