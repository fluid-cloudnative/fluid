package base

import (
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"k8s.io/apimachinery/pkg/types"
)

func (t *TemplateEngine) loggingErrorExceptConflict(err error, message string) error {
	return utils.LoggingErrorExceptConflict(t.Log,
		err,
		message,
		types.NamespacedName{
			Namespace: t.Context.Namespace,
			Name:      t.Context.Name,
		})
}
