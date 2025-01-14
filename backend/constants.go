package main

const (
	host     = "localhost"
	user     = "postgres"
	password = "Egr295_psql"
	dbName   = "wg_api"
)

const (
	MTU       = "1342"
	DNS       = "1.1.1.1"
	AllowedIP = "0.0.0.0/0"
	Endpoint  = "45.142.214.132:51820"
)

const (
	PubKeyPath     = "/etc/wireguard/serverPublicKey"
	PrivateKeyPath = "/etc/wireguard/serverPrivateKey"
)

var (
	token         string
	preExpiredMsg string
	expiredMsg    string
	preDeadMsg    string
	deadMsg       string
)
