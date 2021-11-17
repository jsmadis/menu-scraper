package pkg

import (
	"gopkg.in/yaml.v2"
	"os"
	"path/filepath"
)

type Restaurants struct {
	Restaurants []*RestaurantConfig `yaml:"restaurants"`
}

type RestaurantConfig struct {
	Name string `yaml:"name"`
	Url  string `yaml:"url"`
}

func (r *Restaurants) Load(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	fileData, err := os.ReadFile(absPath)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(fileData, r)
	if err != nil {
		return err
	}

	return nil
}
