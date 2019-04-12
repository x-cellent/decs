package nsq

type config struct {
	nsqdTcpAddress          string
	nsqdHttpAddress         string
	nsqLookupdHttpAddresses []string
}

func newConfig(nsqdTcpAddress, nsqdHttpAddress string, nsqLookupdHttpAddresses ...string) *config {
	return &config{
		nsqdTcpAddress:          nsqdTcpAddress,
		nsqdHttpAddress:         nsqdHttpAddress,
		nsqLookupdHttpAddresses: nsqLookupdHttpAddresses,
	}
}
