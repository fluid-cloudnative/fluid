/* ==================================================================
* Copyright (c) 2023,11.5.
* All rights reserved.
*
* Redistribution and use in source and binary forms, with or without
* modification, are permitted provided that the following conditions
* are met:
*
* 1. Redistributions of source code must retain the above copyright
* notice, this list of conditions and the following disclaimer.
* 2. Redistributions in binary form must reproduce the above copyright
* notice, this list of conditions and the following disclaimer in the
* documentation and/or other materials provided with the
* distribution.
* 3. All advertising materials mentioning features or use of this software
* must display the following acknowledgement:
* This product includes software developed by the xxx Group. and
* its contributors.
* 4. Neither the name of the Group nor the names of its contributors may
* be used to endorse or promote products derived from this software
* without specific prior written permission.
*
* THIS SOFTWARE IS PROVIDED BY xxx,GROUP AND CONTRIBUTORS
* ===================================================================
* Author: xiao shi jie.
*/

package recover

import (
	"context"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/dataset/volume"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubelet"
	"github.com/fluid-cloudnative/fluid/pkg/utils/mountinfo"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	k8sexec "k8s.io/utils/exec"
	"k8s.io/utils/mount"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

const (
	defaultKubeletTimeout          = 10
	defaultFuseRecoveryPeriod      = 5 * time.Second
	defaultRecoverWarningThreshold = 50
	serviceAccountTokenFile        = "/var/run/secrets/kubernetes.io/serviceaccount/token"
	FuseRecoveryPeriod             = "RECOVER_FUSE_PERIOD"
	RecoverWarningThreshold        = "REVOCER_WARNING_THRESHOLD"
)

var _ manager.Runnable = &FuseRecover{}

type FuseRecover struct {
	mount.SafeFormatAndMount
	KubeClient client.Client
	ApiReader  client.Reader
	// KubeletClient *kubelet.KubeletClient
	Recorder record.EventRecorder

	recoverFusePeriod       time.Duration
	recoverWarningThreshold int

	locks *utils.VolumeLocks
}

func initializeKubeletClient() (*kubelet.KubeletClient, error) {
	// get CSI sa token
	tokenByte, err := os.ReadFile(serviceAccountTokenFile)
	if err != nil {
		return nil, errors.Wrap(err, "in cluster mode, find token failed")
	}
	token := string(tokenByte)

	glog.V(3).Infoln("start kubelet client")
	nodeIp := os.Getenv("NODE_IP")
	kubeletClientCert := os.Getenv("KUBELET_CLIENT_CERT")
	kubeletClientKey := os.Getenv("KUBELET_CLIENT_KEY")
	var kubeletTimeout int
	if os.Getenv("KUBELET_TIMEOUT") != "" {
		if kubeletTimeout, err = strconv.Atoi(os.Getenv("KUBELET_TIMEOUT")); err != nil {
			return nil, errors.Wrap(err, "got error when parsing kubelet timeout")
		}
	} else {
		kubeletTimeout = defaultKubeletTimeout
	}
	glog.V(3).Infof("get node ip: %s", nodeIp)
	kubeletClient, err := kubelet.NewKubeletClient(&kubelet.KubeletClientConfig{
		Address: nodeIp,
		Port:    10250,
		TLSClientConfig: rest.TLSClientConfig{
			ServerName: "kubelet",
			CertFile:   kubeletClientCert,
			KeyFile:    kubeletClientKey,
		},
		BearerToken: token,
		HTTPTimeout: time.Duration(kubeletTimeout) * time.Second,
	})

	if err != nil {
		return nil, err
	}

	return kubeletClient, nil
}

func NewFuseRecover(kubeClient client.Client, recorder record.EventRecorder, apiReader client.Reader, locks *utils.VolumeLocks) (*FuseRecover, error) {
	glog.V(3).Infoln("start csi recover")
	mountRoot, err := utils.GetMountRoot()
	if err != nil {
		return nil, errors.Wrap(err, "got err when getting mount root")
	}
	glog.V(3).Infof("Get mount root: %s", mountRoot)

	if err != nil {
		return nil, errors.Wrap(err, "got error when creating kubelet client")
	}

	recoverFusePeriod := utils.GetDurationValueFromEnv(FuseRecoveryPeriod, defaultFuseRecoveryPeriod)
	recoverWarningThreshold, found := utils.GetIntValueFromEnv(RecoverWarningThreshold)
	if !found {
		recoverWarningThreshold = defaultRecoverWarningThreshold
	}
	return &FuseRecover{
		SafeFormatAndMount: mount.SafeFormatAndMount{
			Interface: mount.New(""),
			Exec:      k8sexec.New(),
		},
		KubeClient:              kubeClient,
		ApiReader:               apiReader,
		Recorder:                recorder,
		recoverFusePeriod:       recoverFusePeriod,
		recoverWarningThreshold: recoverWarningThreshold,
		locks:                   locks,
	}, nil
}

func (r *FuseRecover) Start(ctx context.Context) error {
	// do recovering at beginning
	// recover set containerStat in memory, it's none when start
	r.recover()
	r.run(wait.NeverStop)

	return nil
}

func (r *FuseRecover) run(stopCh <-chan struct{}) {
	go wait.Until(r.runOnce, r.recoverFusePeriod, stopCh)
	<-stopCh
	glog.V(3).Info("Shutdown CSI recover.")
}

func (r *FuseRecover) runOnce() {
	r.recover()
}

func (r *FuseRecover) recover() {
	brokenMounts, err := mountinfo.GetBrokenMountPoints()
	if err != nil {
		glog.Error(err)
		return
	}

	for _, point := range brokenMounts {
		r.doRecover(point)
	}
}

func (r *FuseRecover) recoverBrokenMount(point mountinfo.MountPoint) (err error) {
	// recovery for each bind mount path
	mountOption := []string{"bind"}
	if point.ReadOnly {
		mountOption = append(mountOption, "ro")
	}

	glog.V(3).Infof("FuseRecovery: Start exec cmd: mount %s %s -o %v \n", point.SourcePath, point.MountPath, mountOption)
	if err := r.Mount(point.SourcePath, point.MountPath, "none", mountOption); err != nil {
		glog.Errorf("FuseRecovery: exec cmd: mount -o bind %s %s with err :%v", point.SourcePath, point.MountPath, err)
	}
	return
}

// check mountpoint count
// umount duplicate mountpoint util 1 avoiding very large mountinfo file.
// don't umount all item, 'mountPropagation' will lose efficacy.
func (r *FuseRecover) umountDuplicate(point mountinfo.MountPoint) {
	for i := point.Count; i > 1; i-- {
		glog.V(3).Infof("FuseRecovery: count: %d, start exec cmd: umount %s", i, point.MountPath)
		if err := r.Unmount(point.MountPath); err != nil {
			glog.Errorf("FuseRecovery: exec cmd: umount %s with err: %v", point.MountPath, err)
		}
	}
}

func (r *FuseRecover) eventRecord(point mountinfo.MountPoint, eventType, eventReason string) {
	namespacedName := point.NamespacedDatasetName
	strs := strings.Split(namespacedName, "-")
	if len(strs) < 2 {
		glog.V(3).Infof("can't parse dataset from namespacedName: %s", namespacedName)
		return
	}
	namespace, datasetName, err := volume.GetNamespacedNameByVolumeId(r.ApiReader, namespacedName)
	if err != nil {
		glog.Errorf("error get namespacedName by volume id %s: %v", namespacedName, err)
		return
	}

	dataset, err := utils.GetDataset(r.KubeClient, datasetName, namespace)
	if err != nil {
		glog.Errorf("error get dataset %s namespace %s: %v", datasetName, namespace, err)
		return
	}
	glog.V(4).Infof("record to dataset: %s, namespace: %s", dataset.Name, dataset.Namespace)
	switch eventReason {
	case common.FuseRecoverSucceed:
		r.Recorder.Eventf(dataset, eventType, eventReason, "Fuse recover %s succeed", point.MountPath)
	case common.FuseRecoverFailed:
		r.Recorder.Eventf(dataset, eventType, eventReason, "Fuse recover %s failed", point.MountPath)
	case common.FuseUmountDuplicate:
		r.Recorder.Eventf(dataset, eventType, eventReason, "Mountpoint %s has been mounted %v times, unmount duplicate mountpoint to avoid large /proc/self/mountinfo file, this may potential make data access connection broken", point.MountPath, point.Count)
	}
}

func (r *FuseRecover) shouldRecover(mountPath string) (should bool, err error) {
	mounter := mount.New("")
	notMount, err := mounter.IsLikelyNotMountPoint(mountPath)
	if os.IsNotExist(err) || (err == nil && notMount) {
		// Perhaps the mountPath has been cleaned up in other goroutine
		return false, nil
	}
	if err != nil && !mount.IsCorruptedMnt(err) {
		// unexpected error
		return false, err
	}

	return true, nil
}

func (r *FuseRecover) doRecover(point mountinfo.MountPoint) {
	if lock := r.locks.TryAcquire(point.MountPath); !lock {
		glog.V(4).Infof("FuseRecovery: fail to acquire lock on path %s, skip recovering it", point.MountPath)
		return
	}
	defer r.locks.Release(point.MountPath)

	should, err := r.shouldRecover(point.MountPath)
	if err != nil {
		glog.Warningf("FuseRecovery: found path %s which is unable to recover due to error %v, skip it", point.MountPath, err)
		return
	}

	if !should {
		glog.V(3).Infof("FuseRecovery: path %s has already been cleaned up, skip recovering it", point.MountPath)
		return
	}

	glog.V(3).Infof("FuseRecovery: recovering broken mount point: %v", point)
	// if app container restart, umount duplicate mount may lead to recover successed but can not access data
	// so we only umountDuplicate when it has mounted more than the recoverWarningThreshold
	// please refer to https://github.com/fluid-cloudnative/fluid/issues/3399 for more information
	if point.Count > r.recoverWarningThreshold {
		glog.Warningf("FuseRecovery: Mountpoint %s has been mounted %v times, exceeding the recoveryWarningThreshold %v, unmount duplicate mountpoint to avoid large /proc/self/mountinfo file, this may potentially make data access connection broken", point.MountPath, point.Count, r.recoverWarningThreshold)
		r.eventRecord(point, corev1.EventTypeWarning, common.FuseUmountDuplicate)
		r.umountDuplicate(point)
	}
	if err := r.recoverBrokenMount(point); err != nil {
		r.eventRecord(point, corev1.EventTypeWarning, common.FuseRecoverFailed)
		return
	}
	r.eventRecord(point, corev1.EventTypeNormal, common.FuseRecoverSucceed)
}
