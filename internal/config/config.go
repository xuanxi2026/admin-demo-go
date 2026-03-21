package config

type Config struct {
	App     AppConfig     `mapstructure:"app"`
	MySQL   MySQLConfig   `mapstructure:"mysql"`
	Redis   RedisConfig   `mapstructure:"redis"`
	NSQ     NSQConfig     `mapstructure:"nsq"`
	Storage StorageConfig `mapstructure:"storage"`
}

type AppConfig struct {
	Name           string `mapstructure:"name"`
	Mode           string `mapstructure:"mode"`
	Port           int    `mapstructure:"port"`
	JWTSecret      string `mapstructure:"jwt_secret"`
	JWTExpireHours int    `mapstructure:"jwt_expire_hours"`
}

type MySQLConfig struct {
	DSN                    string `mapstructure:"dsn"`
	Database               string `mapstructure:"database"`
	MaxIdleConns           int    `mapstructure:"max_idle_conns"`
	MaxOpenConns           int    `mapstructure:"max_open_conns"`
	ConnMaxLifetimeMinutes int    `mapstructure:"conn_max_lifetime_minutes"`
}

type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type NSQConfig struct {
	ProducerAddr string `mapstructure:"producer_addr"`
	ConsumerAddr string `mapstructure:"consumer_addr"`
	Topic        string `mapstructure:"topic"`
	Channel      string `mapstructure:"channel"`
}

type StorageConfig struct {
	Mode   string            `mapstructure:"mode"` // local / aws-s3 / minio
	Local  StorageLocal      `mapstructure:"local"`
	S3     StorageS3         `mapstructure:"s3"`
	MinIO  StorageMinIO      `mapstructure:"minio"`
	Public StoragePublicLink `mapstructure:"public"`
}

type StorageLocal struct {
	BaseDir string `mapstructure:"base_dir"`
}

type StorageS3 struct {
	Endpoint      string `mapstructure:"endpoint"`
	AccessKey     string `mapstructure:"access_key"`
	SecretKey     string `mapstructure:"secret_key"`
	UseSSL        bool   `mapstructure:"use_ssl"`
	PublicBucket  string `mapstructure:"public_bucket"`
	PrivateBucket string `mapstructure:"private_bucket"`
}

type StorageMinIO struct {
	Endpoint      string `mapstructure:"endpoint"`
	AccessKey     string `mapstructure:"access_key"`
	SecretKey     string `mapstructure:"secret_key"`
	UseSSL        bool   `mapstructure:"use_ssl"`
	PublicBucket  string `mapstructure:"public_bucket"`
	PrivateBucket string `mapstructure:"private_bucket"`
}

type StoragePublicLink struct {
	BaseURL string `mapstructure:"base_url"`
}
