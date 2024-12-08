package simulator

const (
	StateContextKey  = "CTX_State"
	HubsContextKey   = "CTX_Hubs"
	DirectContextKey = "CTX_Direct"
)

type ContextKey struct {
	Key string
}

func GetContextKey(key string) ContextKey {
	return ContextKey{Key: key}
}
