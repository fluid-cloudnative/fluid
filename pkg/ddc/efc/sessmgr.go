package efc

import (
	"context"
	"github.com/fluid-cloudnative/fluid/pkg/utils/docker"
	"os"
	"reflect"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	utilpointer "k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

type config struct {
	SessMgrImage   string
	InitFuseImage  string
	UpdateStrategy appsv1.DaemonSetUpdateStrategyType
}

var _ manager.Runnable = &SessMgrInitializer{}
var _ manager.LeaderElectionRunnable = &SessMgrInitializer{}

type SessMgrInitializer struct {
	client client.Client
}

func NewSessMgrInitializer(client client.Client) *SessMgrInitializer {
	return &SessMgrInitializer{
		client: client,
	}
}

func (s *SessMgrInitializer) initSessMgr(ctx context.Context) error {
	config, err := s.loadSessMgrConfig()
	if err != nil {
		return errors.Wrapf(err, "fail to load sess mgr config from env variables")
	}

	err = s.deploySessMgr(ctx, config)
	if err != nil {
		return errors.Wrapf(err, "fail to deploy sess mgr")
	}
	return nil
}

func (s *SessMgrInitializer) loadSessMgrConfig() (config config, err error) {
	if imageEnvVar, exists := os.LookupEnv(common.EFCSessMgrImageEnv); exists {
		config.SessMgrImage = imageEnvVar
	} else {
		config.SessMgrImage = common.DefaultEFCSessMgrImage
	}

	if imageEnvVar, exists := os.LookupEnv(common.EFCInitFuseImageEnv); exists {
		config.InitFuseImage = imageEnvVar
	} else {
		config.InitFuseImage = common.DefaultEFCInitFuseImage
	}

	if updateStrategyEnvVar, exists := os.LookupEnv(common.EFCSessMgrUpdateStrategyEnv); exists {
		switch updateStrategyEnvVar {
		case string(appsv1.RollingUpdateDaemonSetStrategyType):
			config.UpdateStrategy = appsv1.RollingUpdateDaemonSetStrategyType
		case string(appsv1.OnDeleteDaemonSetStrategyType):
			config.UpdateStrategy = appsv1.OnDeleteDaemonSetStrategyType
		default:
			err = errors.New("Update strategy of SessMgr daemonset must be one of RollingUpdate or OnDelete")
			return
		}
	}

	return
}

func (s *SessMgrInitializer) deploySessMgr(ctx context.Context, config config) error {
	var (
		ns       = &corev1.Namespace{}
		nsExists = true

		ds       = &appsv1.DaemonSet{}
		dsExists = true
	)

	if err := s.client.Get(ctx, types.NamespacedName{Name: common.SessMgrNamespace}, ns); err != nil {
		if utils.IgnoreNotFound(err) != nil {
			return errors.Wrapf(err, "fail to get namespace %s before deploy sessmgr", common.SessMgrNamespace)
		}
		nsExists = false
	}

	if !nsExists {
		ns = &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: common.SessMgrNamespace,
			},
		}
		if err := s.client.Create(ctx, ns); err != nil {
			if utils.IgnoreAlreadyExists(err) != nil {
				return errors.Wrapf(err, "fail to create namespace %s before deploy sessmgr", common.SessMgrNamespace)
			}
		}
	}

	if err := s.client.Get(ctx, types.NamespacedName{Namespace: common.SessMgrNamespace, Name: common.SessMgrDaemonSetName}, ds); err != nil {
		if utils.IgnoreNotFound(err) != nil {
			return errors.Wrapf(err, "fail to get daemonset [%s/%s] before deploy sessmgr", common.SessMgrNamespace, common.SessMgrDaemonSetName)
		}
		dsExists = false
	}

	// Create or update daemonset
	if !dsExists {
		efcSockVolumeType := corev1.HostPathDirectoryOrCreate

		dsToCreate := &appsv1.DaemonSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      common.SessMgrDaemonSetName,
				Namespace: common.SessMgrNamespace,
			},
			Spec: appsv1.DaemonSetSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"app": "efc-sessmgr",
					},
				},
				UpdateStrategy: appsv1.DaemonSetUpdateStrategy{
					Type: config.UpdateStrategy,
				},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							"sidecar.istio.io/inject": "false",
						},
						Labels: map[string]string{
							"app": "efc-sessmgr",
						},
					},
					Spec: corev1.PodSpec{
						HostNetwork:       true,
						PriorityClassName: "system-node-critical",
						NodeSelector: map[string]string{
							common.SessMgrNodeSelectorKey: "true",
						},
						Tolerations: []corev1.Toleration{
							{
								Operator: corev1.TolerationOpExists,
							},
						},
						Affinity: &corev1.Affinity{
							NodeAffinity: &corev1.NodeAffinity{
								RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
									NodeSelectorTerms: []corev1.NodeSelectorTerm{
										{
											MatchExpressions: []corev1.NodeSelectorRequirement{
												{
													Key:      "type",
													Operator: corev1.NodeSelectorOpNotIn,
													Values:   []string{"virtual-kubelet"},
												},
											},
										},
									},
								},
							},
						},
						Volumes: []corev1.Volume{
							corev1.Volume{
								Name: "efc-sock",
								VolumeSource: corev1.VolumeSource{
									HostPath: &corev1.HostPathVolumeSource{
										// TODO: /runtime-mnt to be configurable
										Path: "/runtime-mnt/efc-sock",
										Type: &efcSockVolumeType,
									},
								},
							},
						},
						InitContainers: []corev1.Container{
							corev1.Container{
								Name:    "init-fuse",
								Image:   config.InitFuseImage,
								Command: []string{"/entrypoint.sh"},
								Args: []string{
									"init_fuse",
									"false",
									"none",
								},
								SecurityContext: &corev1.SecurityContext{
									Privileged: utilpointer.Bool(true),
								},
							},
						},
						Containers: []corev1.Container{
							corev1.Container{
								Name:    "sessmgr",
								Command: []string{"/entrypoint.sh"},
								Image:   config.SessMgrImage,
								Args:    []string{"sessmgr"},
								SecurityContext: &corev1.SecurityContext{
									Privileged: utilpointer.BoolPtr(false),
								},
								Lifecycle: &corev1.Lifecycle{
									PreStop: &corev1.LifecycleHandler{
										Exec: &corev1.ExecAction{
											Command: []string{"sh", "-c", "/entrypoint.sh stop_sessmgr"},
										},
									},
								},
								VolumeMounts: []corev1.VolumeMount{
									corev1.VolumeMount{
										MountPath: "/var/run/efc",
										Name:      "efc-sock",
									},
								},
							},
						},
						ImagePullSecrets: docker.GetImagePullSecretsFromEnv(common.EnvImagePullSecretsKey),
					},
				},
			},
		}

		if err := s.client.Create(ctx, dsToCreate); err != nil {
			if utils.IgnoreAlreadyExists(err) != nil {
				return errors.Wrapf(err, "fail to create daemonset [%s/%s]", common.SessMgrNamespace, common.SessMgrDaemonSetName)
			}
		}
	} else {
		err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
			if err := s.client.Get(ctx, types.NamespacedName{Namespace: common.SessMgrNamespace, Name: common.SessMgrDaemonSetName}, ds); err != nil {
				return err
			}

			dsToUpdate := ds.DeepCopy()
			dsToUpdate.Spec.UpdateStrategy.Type = config.UpdateStrategy
			if len(dsToUpdate.Spec.Template.Spec.Containers) != 0 {
				// sessmgr container
				dsToUpdate.Spec.Template.Spec.Containers[0].Image = config.SessMgrImage
			}
			if len(dsToUpdate.Spec.Template.Spec.InitContainers) != 0 {
				// init-fuse init container
				dsToUpdate.Spec.Template.Spec.InitContainers[0].Image = config.InitFuseImage
			}

			if !reflect.DeepEqual(ds, dsToUpdate) {
				if err := s.client.Update(ctx, dsToUpdate); err != nil {
					return err
				}
			}

			return nil
		})

		if err != nil {
			return err
		}
	}

	return nil
}

func (s *SessMgrInitializer) Start(ctx context.Context) error {
	if err := s.initSessMgr(ctx); err != nil {
		return err
	}
	<-ctx.Done()
	return nil
}

func (s *SessMgrInitializer) NeedLeaderElection() bool {
	return true
}
