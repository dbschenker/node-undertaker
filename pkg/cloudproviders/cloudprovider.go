package cloudproviders

type CloudProvider interface {
	ValidateConfig() error
}
