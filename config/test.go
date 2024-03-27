package config

type AppConfig struct {
	DB   DBConfig `yaml:"db"`
	Port int      `yaml:"port"`
}

type DBConfig struct{}

type ConfigManager[T any] struct {
	filename string
	data     T
}

func NewConfigManager[T any](fileName string) *ConfigManager[AppConfig] {
	c := &ConfigManager[AppConfig]{}
	c.load()
	return c
}

func (c *ConfigManager[T]) load() {

}

func (c *ConfigManager[T]) GetConfig() {

}
