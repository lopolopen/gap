package internal

type pair struct {
	Key   string
	Value string
}

func Pair(key string, value string) pair {
	return pair{
		Key:   key,
		Value: value,
	}
}
