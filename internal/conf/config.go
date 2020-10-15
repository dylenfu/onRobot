package conf

var (
	Version            string
	WalletPath         string
	TransferWalletPath string
	ParamsFileDir      string
	DefConf *Config
)

type Config struct {

}

func Init(path string) {

}

type SDKConfig struct {
	JsonRpcAddress   string
	RestfulAddress   string
	WebSocketAddress string

	//Gas Price of transaction
	GasPrice uint64
	//Gas Limit of invoke transaction
	GasLimit uint64
	//Gas Limit of deploy transaction
	GasDeployLimit uint64
}

func SetParamsDir(path string) {
	ParamsFileDir = path
}

func SetWalletPath(path string) {
	WalletPath = path
}

func SetTransferWalletPath(path string) {
	TransferWalletPath = path
}
