package etcd

type Config struct {
	Address []string `json:"address,omitempty" yaml:"address,omitempty" xml:"address,omitempty" bson:"address,omitempty" toml:"address,omitempty" ini:"address,omitempty" env:"ETCD_ADDRESS"`

	Username string `json:"username,omitempty" yaml:"username,omitempty" toml:"username,omitempty" env:"ETCD_USERNAME"`
	Password string `json:"password,omitempty" yaml:"password,omitempty" toml:"password,omitempty" env:"ETCD_PASSWORD"`
}
