/*
Copyright 2017 The Kubernetes Authors.
Copyright 2024 The Fluid Authors.

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

// Package controllerhistory provides utilities for managing ControllerRevisions.
// Adapted from k8s.io/kubernetes/pkg/controller/history.
package controllerhistory

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"sort"
	"strconv"

	"github.com/fluid-cloudnative/fluid/pkg/controllers/workload/v1alpha1/utils/kubecontroller"
	apps "k8s.io/api/apps/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/rand"
	clientset "k8s.io/client-go/kubernetes"
	appslisters "k8s.io/client-go/listers/apps/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/retry"
	hashutil "k8s.io/kubernetes/pkg/util/hash"
)

// ControllerRevisionHashLabel is the label used to indicate the hash value of a ControllerRevision's Data.
const ControllerRevisionHashLabel = "controller.kubernetes.io/hash"

// ControllerRevisionName returns the Name for a ControllerRevision in the form prefix-hash.
func ControllerRevisionName(prefix string, hash string) string {
	if len(prefix) > 223 {
		prefix = prefix[:223]
	}
	return fmt.Sprintf("%s-%s", prefix, hash)
}

// NewControllerRevision returns a ControllerRevision with a ControllerRef pointing to parent.
func NewControllerRevision(parent metav1.Object,
	parentKind schema.GroupVersionKind,
	templateLabels map[string]string,
	data runtime.RawExtension,
	revision int64,
	collisionCount *int32) (*apps.ControllerRevision, error) {
	labelMap := make(map[string]string)
	for k, v := range templateLabels {
		labelMap[k] = v
	}
	cr := &apps.ControllerRevision{
		ObjectMeta: metav1.ObjectMeta{
			Labels:          labelMap,
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(parent, parentKind)},
		},
		Data:     data,
		Revision: revision,
	}
	hash := HashControllerRevision(cr, collisionCount)
	cr.Name = ControllerRevisionName(parent.GetName(), hash)
	cr.Labels[ControllerRevisionHashLabel] = hash
	return cr, nil
}

// HashControllerRevision hashes the contents of revision's Data using FNV hashing.
func HashControllerRevision(revision *apps.ControllerRevision, probe *int32) string {
	hf := fnv.New32()
	if len(revision.Data.Raw) > 0 {
		hf.Write(revision.Data.Raw)
	}
	if revision.Data.Object != nil {
		hashutil.DeepHashObject(hf, revision.Data.Object)
	}
	if probe != nil {
		hf.Write([]byte(strconv.FormatInt(int64(*probe), 10)))
	}
	return rand.SafeEncodeString(fmt.Sprint(hf.Sum32()))
}

// SortControllerRevisions sorts revisions by their Revision.
func SortControllerRevisions(revisions []*apps.ControllerRevision) {
	sort.Stable(byRevision(revisions))
}

// EqualRevision returns true if lhs and rhs are equal.
func EqualRevision(lhs *apps.ControllerRevision, rhs *apps.ControllerRevision) bool {
	var lhsHash, rhsHash *uint32
	if lhs == nil || rhs == nil {
		return lhs == rhs
	}
	if hs, found := lhs.Labels[ControllerRevisionHashLabel]; found {
		hash, err := strconv.ParseInt(hs, 10, 32)
		if err == nil {
			lhsHash = new(uint32)
			*lhsHash = uint32(hash)
		}
	}
	if hs, found := rhs.Labels[ControllerRevisionHashLabel]; found {
		hash, err := strconv.ParseInt(hs, 10, 32)
		if err == nil {
			rhsHash = new(uint32)
			*rhsHash = uint32(hash)
		}
	}
	if lhsHash != nil && rhsHash != nil && *lhsHash != *rhsHash {
		return false
	}
	return bytes.Equal(lhs.Data.Raw, rhs.Data.Raw) && apiequality.Semantic.DeepEqual(lhs.Data.Object, rhs.Data.Object)
}

// FindEqualRevisions returns all ControllerRevisions in revisions that are equal to needle.
func FindEqualRevisions(revisions []*apps.ControllerRevision, needle *apps.ControllerRevision) []*apps.ControllerRevision {
	var eq []*apps.ControllerRevision
	for i := range revisions {
		if EqualRevision(revisions[i], needle) {
			eq = append(eq, revisions[i])
		}
	}
	return eq
}

// byRevision implements sort.Interface to allow ControllerRevisions to be sorted by Revision.
type byRevision []*apps.ControllerRevision

func (br byRevision) Len() int { return len(br) }

func (br byRevision) Less(i, j int) bool {
	if br[i].Revision == br[j].Revision {
		if br[j].CreationTimestamp.Equal(&br[i].CreationTimestamp) {
			return br[i].Name < br[j].Name
		}
		return br[j].CreationTimestamp.After(br[i].CreationTimestamp.Time)
	}
	return br[i].Revision < br[j].Revision
}

func (br byRevision) Swap(i, j int) { br[i], br[j] = br[j], br[i] }

// Interface provides an interface for management of a Controller's history.
type Interface interface {
	ListControllerRevisions(parent metav1.Object, selector labels.Selector) ([]*apps.ControllerRevision, error)
	CreateControllerRevision(parent metav1.Object, revision *apps.ControllerRevision, collisionCount *int32) (*apps.ControllerRevision, error)
	DeleteControllerRevision(revision *apps.ControllerRevision) error
	UpdateControllerRevision(revision *apps.ControllerRevision, newRevision int64) (*apps.ControllerRevision, error)
	AdoptControllerRevision(parent metav1.Object, parentKind schema.GroupVersionKind, revision *apps.ControllerRevision) (*apps.ControllerRevision, error)
	ReleaseControllerRevision(parent metav1.Object, revision *apps.ControllerRevision) (*apps.ControllerRevision, error)
}

// NewHistory returns an instance of Interface that uses client to communicate with the API Server.
func NewHistory(client clientset.Interface, lister appslisters.ControllerRevisionLister) Interface {
	return &realHistory{client, lister}
}

// NewFakeHistory returns an instance of Interface for testing.
func NewFakeHistory(indexer cache.Indexer, lister appslisters.ControllerRevisionLister) Interface {
	return &fakeHistory{indexer, lister}
}

type realHistory struct {
	client clientset.Interface
	lister appslisters.ControllerRevisionLister
}

func (rh *realHistory) ListControllerRevisions(parent metav1.Object, selector labels.Selector) ([]*apps.ControllerRevision, error) {
	history, err := rh.lister.ControllerRevisions(parent.GetNamespace()).List(selector)
	if err != nil {
		return nil, err
	}
	var owned []*apps.ControllerRevision
	for i := range history {
		ref := metav1.GetControllerOfNoCopy(history[i])
		if ref == nil || ref.UID == parent.GetUID() {
			owned = append(owned, history[i])
		}
	}
	return owned, err
}

func (rh *realHistory) CreateControllerRevision(parent metav1.Object, revision *apps.ControllerRevision, collisionCount *int32) (*apps.ControllerRevision, error) {
	if collisionCount == nil {
		return nil, fmt.Errorf("collisionCount should not be nil")
	}
	clone := revision.DeepCopy()
	for {
		hash := HashControllerRevision(revision, collisionCount)
		clone.Name = ControllerRevisionName(parent.GetName(), hash)
		ns := parent.GetNamespace()
		created, err := rh.client.AppsV1().ControllerRevisions(ns).Create(context.TODO(), clone, metav1.CreateOptions{})
		if errors.IsAlreadyExists(err) {
			exists, err := rh.client.AppsV1().ControllerRevisions(ns).Get(context.TODO(), clone.Name, metav1.GetOptions{})
			if err != nil {
				return nil, err
			}
			if bytes.Equal(exists.Data.Raw, clone.Data.Raw) {
				return exists, nil
			}
			*collisionCount++
			continue
		}
		return created, err
	}
}

func (rh *realHistory) UpdateControllerRevision(revision *apps.ControllerRevision, newRevision int64) (*apps.ControllerRevision, error) {
	clone := revision.DeepCopy()
	err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		if clone.Revision == newRevision {
			return nil
		}
		clone.Revision = newRevision
		updated, updateErr := rh.client.AppsV1().ControllerRevisions(clone.Namespace).Update(context.TODO(), clone, metav1.UpdateOptions{})
		if updateErr == nil {
			return nil
		}
		if updated != nil {
			clone = updated
		}
		if updated, err := rh.lister.ControllerRevisions(clone.Namespace).Get(clone.Name); err == nil {
			clone = updated.DeepCopy()
		}
		return updateErr
	})
	return clone, err
}

func (rh *realHistory) DeleteControllerRevision(revision *apps.ControllerRevision) error {
	return rh.client.AppsV1().ControllerRevisions(revision.Namespace).Delete(context.TODO(), revision.Name, metav1.DeleteOptions{})
}

type objectForPatch struct {
	Metadata objectMetaForPatch `json:"metadata"`
}

type objectMetaForPatch struct {
	OwnerReferences []metav1.OwnerReference `json:"ownerReferences"`
	UID             types.UID               `json:"uid"`
}

func (rh *realHistory) AdoptControllerRevision(parent metav1.Object, parentKind schema.GroupVersionKind, revision *apps.ControllerRevision) (*apps.ControllerRevision, error) {
	blockOwnerDeletion := true
	isController := true
	if owner := metav1.GetControllerOfNoCopy(revision); owner != nil {
		return nil, fmt.Errorf("attempt to adopt revision owned by %v", owner)
	}
	addControllerPatch := objectForPatch{
		Metadata: objectMetaForPatch{
			UID: revision.UID,
			OwnerReferences: []metav1.OwnerReference{{
				APIVersion:         parentKind.GroupVersion().String(),
				Kind:               parentKind.Kind,
				Name:               parent.GetName(),
				UID:                parent.GetUID(),
				Controller:         &isController,
				BlockOwnerDeletion: &blockOwnerDeletion,
			}},
		},
	}
	patchBytes, err := json.Marshal(&addControllerPatch)
	if err != nil {
		return nil, err
	}
	return rh.client.AppsV1().ControllerRevisions(parent.GetNamespace()).Patch(context.TODO(), revision.GetName(),
		types.StrategicMergePatchType, patchBytes, metav1.PatchOptions{})
}

func (rh *realHistory) ReleaseControllerRevision(parent metav1.Object, revision *apps.ControllerRevision) (*apps.ControllerRevision, error) {
	dataBytes, err := kubecontroller.GenerateDeleteOwnerRefStrategicMergeBytes(revision.UID, []types.UID{parent.GetUID()})
	if err != nil {
		return nil, err
	}
	released, err := rh.client.AppsV1().ControllerRevisions(revision.GetNamespace()).Patch(context.TODO(), revision.GetName(),
		types.StrategicMergePatchType, dataBytes, metav1.PatchOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, nil
		}
		if errors.IsInvalid(err) {
			return nil, nil
		}
	}
	return released, err
}

// fakeHistory is used for testing
type fakeHistory struct {
	indexer cache.Indexer
	lister  appslisters.ControllerRevisionLister
}

func (fh *fakeHistory) ListControllerRevisions(parent metav1.Object, selector labels.Selector) ([]*apps.ControllerRevision, error) {
	history, err := fh.lister.ControllerRevisions(parent.GetNamespace()).List(selector)
	if err != nil {
		return nil, err
	}
	var owned []*apps.ControllerRevision
	for i := range history {
		ref := metav1.GetControllerOf(history[i])
		if ref == nil || ref.UID == parent.GetUID() {
			owned = append(owned, history[i])
		}
	}
	return owned, err
}

func (fh *fakeHistory) addRevision(revision *apps.ControllerRevision) (*apps.ControllerRevision, error) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(revision)
	if err != nil {
		return nil, err
	}
	obj, found, err := fh.indexer.GetByKey(key)
	if err != nil {
		return nil, err
	}
	if found {
		foundRevision := obj.(*apps.ControllerRevision)
		return foundRevision, errors.NewAlreadyExists(apps.Resource("controllerrevision"), revision.Name)
	}
	return revision, fh.indexer.Update(revision)
}

func (fh *fakeHistory) CreateControllerRevision(parent metav1.Object, revision *apps.ControllerRevision, collisionCount *int32) (*apps.ControllerRevision, error) {
	if collisionCount == nil {
		return nil, fmt.Errorf("collisionCount should not be nil")
	}
	clone := revision.DeepCopy()
	clone.Namespace = parent.GetNamespace()
	for {
		hash := HashControllerRevision(revision, collisionCount)
		clone.Name = ControllerRevisionName(parent.GetName(), hash)
		created, err := fh.addRevision(clone)
		if errors.IsAlreadyExists(err) {
			*collisionCount++
			continue
		}
		return created, err
	}
}

func (fh *fakeHistory) DeleteControllerRevision(revision *apps.ControllerRevision) error {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(revision)
	if err != nil {
		return err
	}
	obj, found, err := fh.indexer.GetByKey(key)
	if err != nil {
		return err
	}
	if !found {
		return errors.NewNotFound(apps.Resource("controllerrevisions"), revision.Name)
	}
	return fh.indexer.Delete(obj)
}

func (fh *fakeHistory) UpdateControllerRevision(revision *apps.ControllerRevision, newRevision int64) (*apps.ControllerRevision, error) {
	clone := revision.DeepCopy()
	clone.Revision = newRevision
	return clone, fh.indexer.Update(clone)
}

func (fh *fakeHistory) AdoptControllerRevision(parent metav1.Object, parentKind schema.GroupVersionKind, revision *apps.ControllerRevision) (*apps.ControllerRevision, error) {
	if owner := metav1.GetControllerOf(revision); owner != nil {
		return nil, fmt.Errorf("attempt to adopt revision owned by %v", owner)
	}
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(revision)
	if err != nil {
		return nil, err
	}
	_, found, err := fh.indexer.GetByKey(key)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, errors.NewNotFound(apps.Resource("controllerrevisions"), revision.Name)
	}
	clone := revision.DeepCopy()
	clone.OwnerReferences = append(clone.OwnerReferences, *metav1.NewControllerRef(parent, parentKind))
	return clone, fh.indexer.Update(clone)
}

func (fh *fakeHistory) ReleaseControllerRevision(parent metav1.Object, revision *apps.ControllerRevision) (*apps.ControllerRevision, error) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(revision)
	if err != nil {
		return nil, err
	}
	_, found, err := fh.indexer.GetByKey(key)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, nil
	}
	clone := revision.DeepCopy()
	refs := clone.OwnerReferences
	clone.OwnerReferences = nil
	for i := range refs {
		if refs[i].UID != parent.GetUID() {
			clone.OwnerReferences = append(clone.OwnerReferences, refs[i])
		}
	}
	return clone, fh.indexer.Update(clone)
}
