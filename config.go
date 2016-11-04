package main

import (
	"crypto/sha256"
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/spf13/viper"
)

type userData struct {
	configFile *viper.Viper
	salt       []byte
}

func saltStringToSHA256(salt string) []byte {

	if salt == "" {
		panic("salt value is empty!")
	}

	returnSalt := make([]byte, 32)
	for e := range sha256.Sum256([]byte(salt)) {
		returnSalt = append(returnSalt, byte(e))
	}
	return returnSalt
}

func parseConfig() *userData {
	viper.SetConfigType("yaml")
	viper.SetConfigName("config")
	viper.AddConfigPath("$HOME/.gcloud-crypto")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Sprintf("Fatal error config file: %s \n", err))
	}

	switch {
	case viper.GetString("bucket") == "":
		panic("'bucket' not set in config file.")
	case viper.GetString("project_id") == "":
		panic("'project_id' not set in config file.")
	}

	// be default, use the SHA256 of the project_id as the salt
	viper.SetDefault("salt", viper.GetString("project_id"))

	saltString := viper.GetString("salt")
	salt := []byte(saltStringToSHA256(saltString))

	log.WithFields(logrus.Fields{"bucket": viper.GetString("bucket"), "project_id": viper.GetString("project_id")}).Debug("Loaded config")
	return &userData{viper.GetViper(), salt}
}
