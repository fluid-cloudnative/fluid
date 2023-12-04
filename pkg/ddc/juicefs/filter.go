package juicefs

// includeEncryptEnvOptionsWithKeys is a function that filters the provided
// slice of EncryptEnvOption instances, including only those options
// whose Name appears in the provided slice of keys. This is useful
// when you want to restrict an operation to a subset of the available
// options.
func includeEncryptEnvOptionsWithKeys(encriptOptions []EncryptEnvOption, keys []string) (result []EncryptEnvOption) {
	encriptOptionsMap := map[string]EncryptEnvOption{}
	for _, opt := range encriptOptions {
		encriptOptionsMap[opt.Name] = opt
	}

	for _, key := range keys {
		if val, found := encriptOptionsMap[key]; found {
			result = append(result, val)
		}
	}

	return
}

// includeOptionsWithKeys is a function that filters the provided
// map of options, including only those entries whose key appears
// in the provided slice of keys. This is useful when you want to
// restrict an operation to a subset of the available options.
func includeOptionsWithKeys(options map[string]string, keys []string) (result map[string]string) {
	result = make(map[string]string)
	for _, key := range keys {
		if val, found := options[key]; found {
			result[key] = val
		}
	}
	return
}
