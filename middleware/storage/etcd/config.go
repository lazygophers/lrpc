package etcd

type Config struct {
	Address []string `json:"address,omitempty" yaml:"address,omitempty" xml:"address,omitempty" bson:"address,omitempty" toml:"address,omitempty" ini:"address,omitempty"`
}
