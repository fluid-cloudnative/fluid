package e2etest

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

var _ = Describe("Test alluxioruntime controller patch node labels", func() {

	const (
		timeout  = time.Second * 30
		interval = time.Millisecond * 250
	)

	ctx := context.TODO()
	var namespace string
	var nodeName string = "fluid-dev-control-plane"
	var fluidOldVal string

	BeforeEach(func() {
		By("Create namespace for testing")
		namespace = randomNamespaceName("patch-node-label")
		ns := v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: namespace,
			},
		}
		Expect(k8sClient.Create(ctx, &ns)).Should(Succeed())

		createdNamespace := v1.Namespace{}
		Eventually(func() error {
			namespaceLookupKey := types.NamespacedName{
				Name: namespace,
			}
			err := k8sClient.Get(ctx, namespaceLookupKey, &createdNamespace)
			if err != nil {
				return err
			}
			return nil
		}, timeout, interval).Should(BeNil())
	})

	AfterEach(func() {
		By("Clean up resources after testing")
		Expect(k8sClient.Delete(ctx, &v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: namespace,
			},
		})).Should(Succeed())

		nodeToCheck := v1.Node{}
		nodeLookupKey := types.NamespacedName{
			Name: nodeName,
		}
		Expect(k8sClient.Get(ctx, nodeLookupKey, &nodeToCheck)).Should(BeNil())
		if fluidOldVal != "" {
			nodeToCheck.Labels["fluid"] = fluidOldVal
		} else {
			delete(nodeToCheck.Labels, "fluid")
		}
		Expect(k8sClient.Update(ctx, &nodeToCheck)).Should(Succeed())
		Eventually(func() error {
			updatedNode := v1.Node{}
			err := k8sClient.Get(ctx, nodeLookupKey, &updatedNode)
			if err != nil {
				return err
			}
			if _, exist := updatedNode.Labels["fluid"]; !(exist == false && fluidOldVal == "" || exist == true && fluidOldVal != "") {
				return fmt.Errorf("fail to delete the label fluid")
			}
			return nil
		}, timeout, interval).Should(BeNil())
	})

	createDataset := func(fileName string) datav1alpha1.Dataset {
		dataset := datav1alpha1.Dataset{}
		Expect(readFile(fileName, &dataset)).Should(BeNil())
		dataset.ObjectMeta.Namespace = namespace
		Expect(k8sClient.Create(ctx, &dataset)).Should(BeNil())
		Eventually(func() error {
			datasetLookupKey := types.NamespacedName{
				Name:      dataset.Name,
				Namespace: dataset.Namespace,
			}
			err := k8sClient.Get(ctx, datasetLookupKey, &dataset)
			if err != nil {
				return err
			}
			return nil
		}, timeout, interval).Should(BeNil())
		return dataset
	}

	createAlluxioruntime := func(fileName string) {
		alluxioruntime := datav1alpha1.AlluxioRuntime{}
		Expect(readFile(fileName, &alluxioruntime)).Should(BeNil())
		alluxioruntime.Namespace = namespace
		Expect(k8sClient.Create(ctx, &alluxioruntime)).Should(BeNil())
		Eventually(func() error {
			alluxioruntimeLookupKey := types.NamespacedName{
				Name:      alluxioruntime.Name,
				Namespace: alluxioruntime.Namespace,
			}
			createdAlluxioruntime := datav1alpha1.AlluxioRuntime{}
			err := k8sClient.Get(ctx, alluxioruntimeLookupKey, &createdAlluxioruntime)
			if err != nil {
				return err
			}
			return nil
		}, timeout, interval).Should(BeNil())
	}

	It("Patch Node", func() {
		By("Add label to Node")
		nodeList := v1.NodeList{}
		Expect(k8sClient.List(ctx, &nodeList)).Should(Succeed())
		Expect(len(nodeList.Items)).ShouldNot(Equal("0"))
		nodeToSchedule := nodeList.Items[0]
		nodeName = nodeToSchedule.Name
		if value, exist := nodeToSchedule.Labels["fluid"]; exist == true && value != "multi-dataset" {
			fluidOldVal = value
		}
		nodeToSchedule.Labels["fluid"] = "multi-dataset"
		Expect(k8sClient.Update(ctx, &nodeToSchedule)).Should(Succeed())
		Eventually(func() error {
			nodeLookupKey := types.NamespacedName{
				Name: nodeName,
			}
			updatedNode := v1.Node{}
			err := k8sClient.Get(ctx, nodeLookupKey, &updatedNode)
			if err != nil {
				return err
			}
			Expect(updatedNode.Labels["fluid"]).Should(Equal("multi-dataset"))
			return nil
		}, timeout, interval).Should(BeNil())

		By("Create dataset and alluxioruntime")
		dataset := createDataset("testdata/dataset-1.yaml")
		createAlluxioruntime("testdata/alluxioruntime-1.yaml")

		// check dataset and runtime bounded
		Eventually(func() error {
			datasetLookupKey := types.NamespacedName{
				Name:      dataset.Name,
				Namespace: dataset.Namespace,
			}
			err := k8sClient.Get(ctx, datasetLookupKey, &dataset)
			if err != nil {
				return err
			}
			if dataset.Status.Phase == datav1alpha1.BoundDatasetPhase {
				return nil
			}
			return err
		}, timeout*3, interval).Should(BeNil())

		time.Sleep(timeout * 3)
		By("Check node labels")
		nodeLookupKey := types.NamespacedName{
			Name: nodeName,
		}
		nodeToCheck := v1.Node{}
		Expect(k8sClient.Get(ctx, nodeLookupKey, &nodeToCheck)).Should(BeNil())
		Expect(nodeToCheck.Labels["fluid.io/dataset-num"]).Should(Equal("1"))

		By("Delete dataset")
		Expect(k8sClient.Delete(ctx, &dataset)).Should(BeNil())

		Eventually(func() error {
			podList := v1.PodList{}
			options := []client.ListOption{
				client.InNamespace(namespace),
			}
			err := k8sClient.List(ctx, &podList, options...)
			if err != nil {
				return err
			}
			if len(podList.Items) != 0 {
				return fmt.Errorf("fail to delete dataset")
			}
			return nil
		}, timeout*2, interval).Should(BeNil())

		time.Sleep(timeout * 2)
		By("Check node labels")
		nodeLookupKey = types.NamespacedName{
			Name: nodeName,
		}
		nodeToCheck = v1.Node{}
		Expect(k8sClient.Get(ctx, nodeLookupKey, &nodeToCheck)).Should(BeNil())
		_, exist := nodeToCheck.Labels["fluid.io/dataset-num"]
		Expect(exist).Should(Equal(false))
	})

	It("patch node label concurrency", func() {
		By("Add label to Node")
		nodeList := v1.NodeList{}
		Expect(k8sClient.List(ctx, &nodeList)).Should(Succeed())
		Expect(len(nodeList.Items)).ShouldNot(Equal("0"))
		nodeToSchedule := nodeList.Items[0]
		nodeName = nodeToSchedule.Name
		if value, exist := nodeToSchedule.Labels["fluid"]; exist == true && value != "multi-dataset" {
			fluidOldVal = value
		}
		nodeToSchedule.Labels["fluid"] = "multi-dataset"
		Expect(k8sClient.Update(ctx, &nodeToSchedule)).Should(Succeed())
		Eventually(func() error {
			nodeLookupKey := types.NamespacedName{
				Name: nodeName,
			}
			updatedNode := v1.Node{}
			err := k8sClient.Get(ctx, nodeLookupKey, &updatedNode)
			if err != nil {
				return err
			}
			Expect(updatedNode.Labels["fluid"]).Should(Equal("multi-dataset"))
			return nil
		}, timeout, interval).Should(BeNil())

		// add datasets concurrency
		for i := 1; i <= 3; i++ {
			go func(index int) {
				defer GinkgoRecover()
				datasetFileName := "testdata/dataset-" + strconv.Itoa(index) + ".yaml"
				alluxioruntimeFileName := "testdata/alluxioruntime-" + strconv.Itoa(index) + ".yaml"
				dataset := createDataset(datasetFileName)
				createAlluxioruntime(alluxioruntimeFileName)
				// check dataset and runtime bounded
				Eventually(func() error {
					datasetLookupKey := types.NamespacedName{
						Name:      dataset.Name,
						Namespace: dataset.Namespace,
					}
					err := k8sClient.Get(ctx, datasetLookupKey, &dataset)
					if err != nil {
						return err
					}
					if dataset.Status.Phase == datav1alpha1.BoundDatasetPhase {
						return nil
					}
					return err
				}, timeout*3, interval).Should(BeNil())
			}(i)
		}

		time.Sleep(timeout * 4)
		By("Check node labels")
		nodeLookupKey := types.NamespacedName{
			Name: nodeName,
		}
		nodeToCheck := v1.Node{}
		Expect(k8sClient.Get(ctx, nodeLookupKey, &nodeToCheck)).Should(BeNil())
		Expect(nodeToCheck.Labels["fluid.io/dataset-num"]).Should(Equal("3"))

		// add and delete datasets concurrnecy
		for i := 1; i <= 2; i++ {
			go func(index int) {
				defer GinkgoRecover()
				if index == 1 {
					deletedDataset := datav1alpha1.Dataset{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: namespace,
							Name:      "hbase1",
						},
					}
					Expect(k8sClient.Delete(ctx, &deletedDataset)).Should(BeNil())
				} else {
					datasetFileName := "testdata/dataset-4.yaml"
					alluxioruntimeFileName := "testdata/alluxioruntime-4.yaml"
					dataset := createDataset(datasetFileName)
					createAlluxioruntime(alluxioruntimeFileName)

					// check dataset and runtime bounded
					Eventually(func() error {
						datasetLookupKey := types.NamespacedName{
							Name:      dataset.Name,
							Namespace: dataset.Namespace,
						}
						err := k8sClient.Get(ctx, datasetLookupKey, &dataset)
						if err != nil {
							return err
						}
						if dataset.Status.Phase == datav1alpha1.BoundDatasetPhase {
							return nil
						}
						return err
					}, timeout*3, interval).Should(BeNil())
				}
			}(i)
		}

		time.Sleep(timeout * 3)
		By("Check node labels")
		nodeLookupKey = types.NamespacedName{
			Name: nodeName,
		}
		nodeToCheck = v1.Node{}
		Expect(k8sClient.Get(ctx, nodeLookupKey, &nodeToCheck)).Should(BeNil())
		Expect(nodeToCheck.Labels["fluid.io/dataset-num"]).Should(Equal("3"))

		//delete datasets concurrency
		for i := 2; i <= 4; i++ {
			go func(index int) {
				defer GinkgoRecover()
				deletedDataset := datav1alpha1.Dataset{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: namespace,
						Name:      "hbase" + strconv.Itoa(index),
					},
				}
				Expect(k8sClient.Delete(ctx, &deletedDataset)).Should(BeNil())
			}(i)
		}

		Eventually(func() error {
			podList := v1.PodList{}
			options := []client.ListOption{
				client.InNamespace(namespace),
			}
			err := k8sClient.List(ctx, &podList, options...)
			if err != nil {
				return err
			}
			if len(podList.Items) != 0 {
				return fmt.Errorf("fail to delete dataset")
			}
			return nil
		}, timeout*2, interval).Should(BeNil())

		time.Sleep(timeout * 2)
		By("Check node labels")
		nodeLookupKey = types.NamespacedName{
			Name: nodeName,
		}
		nodeToCheck = v1.Node{}
		Expect(k8sClient.Get(ctx, nodeLookupKey, &nodeToCheck)).Should(BeNil())
		_, exist := nodeToCheck.Labels["fluid.io/dataset-num"]
		Expect(exist).Should(Equal(false))
	})

})

// readFile will read a yaml k8s object to runtime.Object.
func readFile(fileName string, object runtime.Object) error {
	data, err := os.ReadFile(fileName)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(data, object)
	return err
}

// randomNamespaceName creates name of namespace randomly.
func randomNamespaceName(basic string) string {
	return fmt.Sprintf("%s-%s", basic, strconv.FormatInt(rand.Int63(), 16))
}
