/*
  Copyright 2023 The Fluid Authors.

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

package base

import (
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
)

func (t *TemplateEngine) MigrateData(ctx cruntime.ReconcileRequestContext, targetDataMigrate datav1alpha1.DataMigrate) (err error) {
	if err = t.Implement.CreateDataMigrateJob(ctx, targetDataMigrate); err != nil {
		t.Log.Error(err, "Failed to create the DataMigrate job.")
		return err
	}
	return err
}
