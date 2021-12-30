// Package flat flattens a nested Golang map into a one level deep map. Flat also supports unflatten, turn a one level map into nested one.
//
// You can flatten a Go map
//
//     in = map[string]interface{}{
//         "foo": map[string]interface{}{
//             "bar": map[string]interface{}{
//                 "t": 123,
//             },
//             "k": 456,
//         },
//     }
//
//     out, err := flat.Flatten(in, nil)
//     // out = map[string]interface{}{
//     //     "foo.bar.t": 123,
//     //     "foo.k": 456,
//     // }
//
// and a reverse with unflatten
//     in = map[string]interface{}{
//         "foo.bar.t": 123,
//         "foo.k": 456,
//     }
//     out, err := flat.Unflatten(in, nil)
//     // out = map[string]interface{}{
//     //     "foo": map[string]interface{}{
//     //         "bar": map[string]interface{}{
//     //             "t": 123,
//     //         },
//     //         "k": 456,
//     //     },
//     // }
//
package flat
