package config

import (
	"io/ioutil"
	"os"
	"path"

	"gopkg.in/yaml.v2"

	"github.com/feeltheajf/ztunnel/x/fs"
)

const (
	App         = "ztunnel"
	DefaultPath = App + ".yml"

	HeaderAPIToken = "X-Api-Token"
)

var (
	DefaultDir = path.Join(fs.UserConfigDir(), App)
)

func init() {
	err := fs.Mkdir(DefaultDir)
	if err != nil {
		panic(err)
	}
}

type Config struct {
	Address string            `yaml:"address"`
	Servers map[string]string `yaml:"servers"`
}

func Load(path string) (*Config, error) {
	cfg := new(Config)
	return cfg, load(path, cfg)
}

func load(path string, i interface{}) error {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	b = []byte(os.ExpandEnv(string(b)))
	if err := yaml.UnmarshalStrict(b, i); err != nil {
		return err
	}

	return nil
}

func Path(filename string) string {
	return path.Join(DefaultDir, filename)
}
