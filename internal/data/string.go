package data

func NotNilStrPtr(str *string) bool {
	return str != nil && *str != ""
}

func GetOrDefaultStrPtr(str *string, def *string) *string {
	if str != nil && *str != "" {
		return str
	}

	if def != nil && *def != "" {
		return def
	}
	return nil
}

func GetOrDefaultStr(str string, def string) string {
	if str != "" {
		return str
	}

	if def != "" {
		return def
	}
	return ""
}
