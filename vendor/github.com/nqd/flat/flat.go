package flat

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/imdario/mergo"
)

// Options the flatten options.
// By default: Demiliter = "."
type Options struct {
	Delimiter string
	Safe      bool
	MaxDepth  int
}

// Flatten the map, it returns a map one level deep
// regardless of how nested the original map was.
// By default, the flatten has Delimiter = ".", and
// no limitation of MaxDepth
func Flatten(nested map[string]interface{}, opts *Options) (m map[string]interface{}, err error) {
	if opts == nil {
		opts = &Options{
			Delimiter: ".",
		}
	}

	m, err = flatten("", 0, nested, opts)

	return
}

func flatten(prefix string, depth int, nested interface{}, opts *Options) (flatmap map[string]interface{}, err error) {
	flatmap = make(map[string]interface{})

	switch nested := nested.(type) {
	case map[string]interface{}:
		if opts.MaxDepth != 0 && depth >= opts.MaxDepth {
			flatmap[prefix] = nested
			return
		}
		if reflect.DeepEqual(nested, map[string]interface{}{}) {
			flatmap[prefix] = nested
			return
		}
		for k, v := range nested {
			// create new key
			newKey := k
			if prefix != "" {
				newKey = prefix + opts.Delimiter + newKey
			}
			fm1, fe := flatten(newKey, depth+1, v, opts)
			if fe != nil {
				err = fe
				return
			}
			update(flatmap, fm1)
		}
	case []interface{}:
		if opts.Safe {
			flatmap[prefix] = nested
			return
		}
		if reflect.DeepEqual(nested, []interface{}{}) {
			flatmap[prefix] = nested
			return
		}
		for i, v := range nested {
			newKey := strconv.Itoa(i)
			if prefix != "" {
				newKey = prefix + opts.Delimiter + newKey
			}
			fm1, fe := flatten(newKey, depth+1, v, opts)
			if fe != nil {
				err = fe
				return
			}
			update(flatmap, fm1)
		}
	default:
		flatmap[prefix] = nested
	}
	return
}

// update is the function that update to map with from
// example:
// to = {"hi": "there"}
// from = {"foo": "bar"}
// then, to = {"hi": "there", "foo": "bar"}
func update(to map[string]interface{}, from map[string]interface{}) {
	for kt, vt := range from {
		to[kt] = vt
	}
}

// Unflatten the map, it returns a nested map of a map
// By default, the flatten has Delimiter = "."
func Unflatten(flat map[string]interface{}, opts *Options) (nested map[string]interface{}, err error) {
	if opts == nil {
		opts = &Options{
			Delimiter: ".",
		}
	}
	nested, err = unflatten(flat, opts)
	return
}

func unflatten(flat map[string]interface{}, opts *Options) (nested map[string]interface{}, err error) {
	nested = make(map[string]interface{})

	for k, v := range flat {
		temp := uf(k, v, opts).(map[string]interface{})
		err = mergo.Merge(&nested, temp, func(c *mergo.Config) { c.Overwrite = true })
		if err != nil {
			return
		}
	}

	return
}

func uf(k string, v interface{}, opts *Options) (n interface{}) {
	n = v

	keys := strings.Split(k, opts.Delimiter)

	for i := len(keys) - 1; i >= 0; i-- {
		temp := make(map[string]interface{})
		temp[keys[i]] = n
		n = temp
	}

	return
}
