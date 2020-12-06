/*
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package utils

import (
	"strings"

	units "github.com/docker/go-units"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TranformQuantityToAlluxioUnit(q *resource.Quantity) (value string) {
	value = q.String()

	if strings.HasSuffix(value, "i") {
		value = strings.ReplaceAll(value, "i", "B")
	}
	return
	// return units.BytesSize(units.BytesSize(float64(q.Value())))

}

func TranformQuantityToUnits(q *resource.Quantity) (value string) {
	// value = q.String()

	// if strings.HasSuffix(value, "i") {
	// 	strings.ReplaceAll(value, "i", "B")
	// }

	return units.BytesSize(float64(q.Value()))

}
