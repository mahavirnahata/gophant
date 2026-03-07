package cache

func tagKey(tag, key string) string {
	if tag == "" {
		return key
	}
	return "tag:" + tag + ":" + key
}
