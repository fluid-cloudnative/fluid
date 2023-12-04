package juicefs

// filter is an interface for filtering options
type filter interface {
	filterOption(options map[string]string) (result map[string]string)
	filterEncryptEnvOptions(encriptOptions []EncryptEnvOption) (result []EncryptEnvOption)
}

// abstractFilter is a concrete filter type. It has fields for allowed and disallowed option keys
type abstractFilter struct {
	allowOptionKey              []string
	allowEncryptEnvOptionKey    []string
	disallowOptionKey           []string
	disallowEncryptEnvOptionKey []string
}

// formatCmdFilter is a struct for filtering format command options. It embeds abstractFilter
type formatCmdFilter struct {
	abstractFilter
}

// filterOption implements the filtering of options. It creates a new options map that includes all allowed keys, and then excludes any disallowed keys
func (f abstractFilter) filterOption(options map[string]string) (result map[string]string) {
	// If allowed option keys are initialized, create a result map that includes only these
	if f.allowOptionKey != nil {
		result = includeOptionsWithKeys(options, f.allowOptionKey)
	}

	// If disallowed option keys are initialized, exclude these from the result map
	if f.disallowOptionKey != nil {
		result = excludeOptionsWithKeys(result, f.disallowOptionKey)
	}

	return
}

// filterEncryptEnvOptions implements the filtering of encrypted environment options. It creates a new options slice that includes all allowed keys, and then excludes any disallowed keys
func (f abstractFilter) filterEncryptEnvOptions(encriptOptions []EncryptEnvOption) (result []EncryptEnvOption) {
	// If allowed option keys are initialized, create a result slice that includes only these
	if f.allowEncryptEnvOptionKey != nil {
		result = includeEncryptEnvOptionsWithKeys(encriptOptions, f.allowEncryptEnvOptionKey)
	}

	// If disallowed option keys are initialized, exclude these from the result slice
	if f.disallowEncryptEnvOptionKey != nil {
		result = excludeEncryptEnvOptionsWithKeys(result, f.disallowEncryptEnvOptionKey)
	}

	return
}

// buildFormatCmdFilterForCommunityEdition builds a command filter for the community edition
func buildFormatCmdFilterForEnterpriseEdition() filter {
	f := formatCmdFilter{}
	f.allowOptionKey = []string{JuiceBucket2}
	f.allowEncryptEnvOptionKey = []string{AccessKey2, SecretKey2}
	return f
}

// buildFormatCmdFilterForCommunityEdition builds a command filter for the community edition
func buildFormatCmdFilterForCommunityEdition() filter {
	f := formatCmdFilter{}
	return f
}

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

// excludeOptionsWithKeys is a function that filters the provided
// map of options, excluding those entries whose key appears
// in the provided slice of keys. This is useful when you want to
// exclude a subset of the available options.
func excludeOptionsWithKeys(options map[string]string, keys []string) (result map[string]string) {
	result = make(map[string]string)
	// Loop through the options map
	for k, v := range options {
		// Flag that indicates if key is in the exclude list
		excludeKey := false
		for _, key := range keys {
			if k == key {
				excludeKey = true
				break
			}
		}
		// If key was not found in the exclude list, add to result
		if !excludeKey {
			result[k] = v
		}
	}
	return
}

// excludeEncryptEnvOptionsWithKeys is a function that filters the provided
// slice of EncryptEnvOption instances, excluding those options
// whose Name appears in the provided slice of keys. This is useful
// when you want to restrict an operation to a subset of the available
// options.
func excludeEncryptEnvOptionsWithKeys(encriptOptions []EncryptEnvOption, keys []string) (result []EncryptEnvOption) {
	// Create a map for checking exclusion keys quickly
	keysMap := make(map[string]bool)
	for _, key := range keys {
		keysMap[key] = true
	}

	// Loop through the encriptOptions slice
	for _, opt := range encriptOptions {
		// If the Name of opt is not in keysMap, add it to the result
		if _, found := keysMap[opt.Name]; !found {
			result = append(result, opt)
		}
	}

	return
}
