package config

type ConfigDB struct {
	PassWord string
	User     string
	Host     string
	Port     string
	DBName   string
}

type ConfigRDB struct {
	Addr     string
	PassWord string
}

type ConfigData struct {
	Database *ConfigDB
	Redis    *ConfigRDB
}
