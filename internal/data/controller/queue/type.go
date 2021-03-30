package queue

import "time"

type StanConnection struct {
	ClusterID          string `yaml:"clusterid"`
	ClientID           string `yaml:"clientid"`
	SubjectPublish     string `yaml:"subjectpublish"`
	MaxPubAcksInFlight int    `yaml:"maxpubacksinflight"`
	GroupName          string `yaml:"groupname"`
	DurableName        string `yaml:"durablename"`
	MaxWithoutAck      int    `yaml:"maxwithoutack"`
	AckWaitTimeSec     int    `yaml:"ackwaittimesec"`
	AckWaitDelay       int    `yaml:"ackwaitdelay"`
}
type NatsConnection struct {
	DataConnection
	TimeoutMS           time.Duration `yaml:"timeoutms"`
	PingIntervalMS      time.Duration `yaml:"pingintervalms"`
	MaxPingsOutstanding int           `yaml:"maxpingsoutstanding"`
	ReconnectWait       time.Duration `yaml:"reconnectwait"`
	ReconnectBufSize    int           `yaml:"reconnectbufsize"`
}
type DataConnection struct {
	Provider string `yaml:"provider"`
	Address  string `yaml:"address"`
	Port     string `yaml:"port"`
	Login    string `yaml:"login"`
	Password string `yaml:"password"`
	Name     string `yaml:"name"`
}