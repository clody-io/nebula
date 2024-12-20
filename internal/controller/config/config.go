package config

var CFG Config

type OpenstackAuthConfig struct {
	Region        string `json:"region"`
	ProjectName   string `json:"projectName"`
	UserDomain    string `json:"userDomain"`
	ProjectDomain string `json:"projectDomain"`
	UserName      string `json:"userName"`
	UserPassword  string `json:"userPassword"`
}

type Config struct {
	OpenstackAuth OpenstackAuthConfig
}

func Init(cfg Config) error {
	CFG = cfg

	return nil
}
