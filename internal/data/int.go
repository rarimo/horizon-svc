package data

func GetOrDefaultUint64Ptr(value *uint64, def *uint64) *uint64 {
	if value != nil && *value != 0 {
		return value
	}

	if def != nil && *def != 0 {
		return def
	}
	return nil
}
