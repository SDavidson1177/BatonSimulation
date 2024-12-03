package simulator

type ContextKey struct {
	Key string
}

func GetContextKey(key string) ContextKey {
	return ContextKey{Key: key}
}
