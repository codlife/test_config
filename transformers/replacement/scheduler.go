/*
Copyright 2014 The Kubernetes Authors.

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

package scheduler

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	appsinformers "k8s.io/client-go/informers/apps/v1"
	coreinformers "k8s.io/client-go/informers/core/v1"
	policyinformers "k8s.io/client-go/informers/policy/v1beta1"
	storageinformers "k8s.io/client-go/informers/storage/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/record"
	"k8s.io/kubernetes/pkg/features"
	"k8s.io/kubernetes/pkg/scheduler/algorithm/predicates"
	schedulerapi "k8s.io/kubernetes/pkg/scheduler/api"
	latestschedulerapi "k8s.io/kubernetes/pkg/scheduler/api/latest"
	kubeschedulerconfig "k8s.io/kubernetes/pkg/scheduler/apis/config"
	"k8s.io/kubernetes/pkg/scheduler/core"
	"k8s.io/kubernetes/pkg/scheduler/factory"
	schedulerinternalcache "k8s.io/kubernetes/pkg/scheduler/internal/cache"
	"k8s.io/kubernetes/pkg/scheduler/metrics"
	"k8s.io/kubernetes/pkg/scheduler/util"
	"k8s.io/client-go/dynamic"
	mtv1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog"
	"strings"
	"encoding/json"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/api/resource"
)

const (
	// BindTimeoutSeconds defines the default bind timeout
	BindTimeoutSeconds = 100
)

// Scheduler watches for new unscheduled pods. It attempts to find
// nodes that they fit on and writes bindings back to the api server.
type Scheduler struct {
	config *factory.Config
}

// Cache returns the cache in scheduler for test to check the data in scheduler.
func (sched *Scheduler) Cache() schedulerinternalcache.Cache {
	return sched.config.SchedulerCache
}

type schedulerOptions struct {
	schedulerName                  string
	hardPodAffinitySymmetricWeight int32
	enableEquivalenceClassCache    bool
	disablePreemption              bool
	percentageOfNodesToScore       int32
	bindTimeoutSeconds             int64
}

// Option configures a Scheduler
type Option func(*schedulerOptions)

// WithName sets schedulerName for Scheduler, the default schedulerName is default-scheduler
func WithName(schedulerName string) Option {
	return func(o *schedulerOptions) {
		o.schedulerName = schedulerName
	}
}

// WithHardPodAffinitySymmetricWeight sets hardPodAffinitySymmetricWeight for Scheduler, the default value is 1
func WithHardPodAffinitySymmetricWeight(hardPodAffinitySymmetricWeight int32) Option {
	return func(o *schedulerOptions) {
		o.hardPodAffinitySymmetricWeight = hardPodAffinitySymmetricWeight
	}
}

// WithEquivalenceClassCacheEnabled sets enableEquivalenceClassCache for Scheduler, the default value is false
func WithEquivalenceClassCacheEnabled(enableEquivalenceClassCache bool) Option {
	return func(o *schedulerOptions) {
		o.enableEquivalenceClassCache = enableEquivalenceClassCache
	}
}

// WithPreemptionDisabled sets disablePreemption for Scheduler, the default value is false
func WithPreemptionDisabled(disablePreemption bool) Option {
	return func(o *schedulerOptions) {
		o.disablePreemption = disablePreemption
	}
}

// WithPercentageOfNodesToScore sets percentageOfNodesToScore for Scheduler, the default value is 50
func WithPercentageOfNodesToScore(percentageOfNodesToScore int32) Option {
	return func(o *schedulerOptions) {
		o.percentageOfNodesToScore = percentageOfNodesToScore
	}
}

// WithBindTimeoutSeconds sets bindTimeoutSeconds for Scheduler, the default value is 100
func WithBindTimeoutSeconds(bindTimeoutSeconds int64) Option {
	return func(o *schedulerOptions) {
		o.bindTimeoutSeconds = bindTimeoutSeconds
	}
}

var defaultSchedulerOptions = schedulerOptions{
	schedulerName:                  v1.DefaultSchedulerName,
	hardPodAffinitySymmetricWeight: v1.DefaultHardPodAffinitySymmetricWeight,
	enableEquivalenceClassCache:    false,
	disablePreemption:              false,
	percentageOfNodesToScore:       schedulerapi.DefaultPercentageOfNodesToScore,
	bindTimeoutSeconds:             BindTimeoutSeconds,
}

// New returns a Scheduler
func New(client clientset.Interface,
	nodeInformer coreinformers.NodeInformer,
	podInformer coreinformers.PodInformer,
	pvInformer coreinformers.PersistentVolumeInformer,
	pvcInformer coreinformers.PersistentVolumeClaimInformer,
	replicationControllerInformer coreinformers.ReplicationControllerInformer,
	replicaSetInformer appsinformers.ReplicaSetInformer,
	statefulSetInformer appsinformers.StatefulSetInformer,
	serviceInformer coreinformers.ServiceInformer,
	pdbInformer policyinformers.PodDisruptionBudgetInformer,
	storageClassInformer storageinformers.StorageClassInformer,
	recorder record.EventRecorder,
	schedulerAlgorithmSource kubeschedulerconfig.SchedulerAlgorithmSource,
	stopCh <-chan struct{},
	opts ...func(o *schedulerOptions)) (*Scheduler, error) {

	options := defaultSchedulerOptions
	for _, opt := range opts {
		opt(&options)
	}

	// Set up the configurator which can create schedulers from configs.
	configurator := factory.NewConfigFactory(&factory.ConfigFactoryArgs{
		SchedulerName:                  options.schedulerName,
		Client:                         client,
		NodeInformer:                   nodeInformer,
		PodInformer:                    podInformer,
		PvInformer:                     pvInformer,
		PvcInformer:                    pvcInformer,
		ReplicationControllerInformer:  replicationControllerInformer,
		ReplicaSetInformer:             replicaSetInformer,
		StatefulSetInformer:            statefulSetInformer,
		ServiceInformer:                serviceInformer,
		PdbInformer:                    pdbInformer,
		StorageClassInformer:           storageClassInformer,
		HardPodAffinitySymmetricWeight: options.hardPodAffinitySymmetricWeight,
		EnableEquivalenceClassCache:    options.enableEquivalenceClassCache,
		DisablePreemption:              options.disablePreemption,
		PercentageOfNodesToScore:       options.percentageOfNodesToScore,
		BindTimeoutSeconds:             options.bindTimeoutSeconds,
	})
	var config *factory.Config
	source := schedulerAlgorithmSource
	switch {
	case source.Provider != nil:
		// Create the config from a named algorithm provider.
		sc, err := configurator.CreateFromProvider(*source.Provider)
		if err != nil {
			return nil, fmt.Errorf("couldn't create scheduler using provider %q: %v", *source.Provider, err)
		}
		config = sc
	case source.Policy != nil:
		// Create the config from a user specified policy source.
		policy := &schedulerapi.Policy{}
		switch {
		case source.Policy.File != nil:
			if err := initPolicyFromFile(source.Policy.File.Path, policy); err != nil {
				return nil, err
			}
		case source.Policy.ConfigMap != nil:
			if err := initPolicyFromConfigMap(client, source.Policy.ConfigMap, policy); err != nil {
				return nil, err
			}
		}
		sc, err := configurator.CreateFromConfig(*policy)
		if err != nil {
			return nil, fmt.Errorf("couldn't create scheduler from policy: %v", err)
		}
		config = sc
	default:
		return nil, fmt.Errorf("unsupported algorithm source: %v", source)
	}
	// Additional tweaks to the config produced by the configurator.
	config.Recorder = recorder
	config.DisablePreemption = options.disablePreemption
	config.StopEverything = stopCh
	// Create the scheduler.
	sched := NewFromConfig(config)
	return sched, nil
}

// initPolicyFromFile initialize policy from file
func initPolicyFromFile(policyFile string, policy *schedulerapi.Policy) error {
	// Use a policy serialized in a file.
	_, err := os.Stat(policyFile)
	if err != nil {
		return fmt.Errorf("missing policy config file %s", policyFile)
	}
	data, err := ioutil.ReadFile(policyFile)
	if err != nil {
		return fmt.Errorf("couldn't read policy config: %v", err)
	}
	err = runtime.DecodeInto(latestschedulerapi.Codec, []byte(data), policy)
	if err != nil {
		return fmt.Errorf("invalid policy: %v", err)
	}
	return nil
}

// initPolicyFromConfigMap initialize policy from configMap
func initPolicyFromConfigMap(client clientset.Interface, policyRef *kubeschedulerconfig.SchedulerPolicyConfigMapSource, policy *schedulerapi.Policy) error {
	// Use a policy serialized in a config map value.
	policyConfigMap, err := client.CoreV1().ConfigMaps(policyRef.Namespace).Get(policyRef.Name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("couldn't get policy config map %s/%s: %v", policyRef.Namespace, policyRef.Name, err)
	}
	data, found := policyConfigMap.Data[kubeschedulerconfig.SchedulerPolicyConfigMapKey]
	if !found {
		return fmt.Errorf("missing policy config map value at key %q", kubeschedulerconfig.SchedulerPolicyConfigMapKey)
	}
	err = runtime.DecodeInto(latestschedulerapi.Codec, []byte(data), policy)
	if err != nil {
		return fmt.Errorf("invalid policy: %v", err)
	}
	return nil
}

// NewFromConfigurator returns a new scheduler that is created entirely by the Configurator.  Assumes Create() is implemented.
// Supports intermediate Config mutation for now if you provide modifier functions which will run after Config is created.
func NewFromConfigurator(c factory.Configurator, modifiers ...func(c *factory.Config)) (*Scheduler, error) {
	cfg, err := c.Create()
	if err != nil {
		return nil, err
	}
	// Mutate it if any functions were provided, changes might be required for certain types of tests (i.e. change the recorder).
	for _, modifier := range modifiers {
		modifier(cfg)
	}
	// From this point on the config is immutable to the outside.
	s := &Scheduler{
		config: cfg,
	}
	metrics.Register()
	return s, nil
}

// NewFromConfig returns a new scheduler using the provided Config.
func NewFromConfig(config *factory.Config) *Scheduler {
	metrics.Register()
	return &Scheduler{
		config: config,
	}
}

// Run begins watching and scheduling. It waits for cache to be synced, then starts a goroutine and returns immediately.
func (sched *Scheduler) Run() {
	if !sched.config.WaitForCacheSync() {
		return
	}

	go wait.Until(sched.scheduleOne, 0, sched.config.StopEverything)
	go transformer()

}

// Config returns scheduler's config pointer. It is exposed for testing purposes.
func (sched *Scheduler) Config() *factory.Config {
	return sched.config
}

// schedule implements the scheduling algorithm and returns the suggested host.
func (sched *Scheduler) schedule(pod *v1.Pod) (string, error) {
	host, err := sched.config.Algorithm.Schedule(pod, sched.config.NodeLister)
	if err != nil {
		pod = pod.DeepCopy()
		sched.config.Error(pod, err)
		sched.config.Recorder.Eventf(pod, v1.EventTypeWarning, "FailedScheduling", "%v", err)
		sched.config.PodConditionUpdater.Update(pod, &v1.PodCondition{
			Type:          v1.PodScheduled,
			Status:        v1.ConditionFalse,
			LastProbeTime: metav1.Now(),
			Reason:        v1.PodReasonUnschedulable,
			Message:       err.Error(),
		})
		return "", err
	}
	return host, err
}

// preempt tries to create room for a pod that has failed to schedule, by preempting lower priority pods if possible.
// If it succeeds, it adds the name of the node where preemption has happened to the pod annotations.
// It returns the node name and an error if any.
func (sched *Scheduler) preempt(preemptor *v1.Pod, scheduleErr error) (string, error) {
	if !util.PodPriorityEnabled() || sched.config.DisablePreemption {
		klog.V(3).Infof("Pod priority feature is not enabled or preemption is disabled by scheduler configuration." +
			" No preemption is performed.")
		return "", nil
	}
	preemptor, err := sched.config.PodPreemptor.GetUpdatedPod(preemptor)
	if err != nil {
		klog.Errorf("Error getting the updated preemptor pod object: %v", err)
		return "", err
	}

	node, victims, nominatedPodsToClear, err := sched.config.Algorithm.Preempt(preemptor, sched.config.NodeLister, scheduleErr)
	metrics.PreemptionVictims.Set(float64(len(victims)))
	if err != nil {
		klog.Errorf("Error preempting victims to make room for %v/%v.", preemptor.Namespace, preemptor.Name)
		return "", err
	}
	var nodeName = ""
	if node != nil {
		nodeName = node.Name
		// Update the scheduling queue with the nominated pod information. Without
		// this, there would be a race condition between the next scheduling cycle
		// and the time the scheduler receives a Pod Update for the nominated pod.
		sched.config.SchedulingQueue.UpdateNominatedPodForNode(preemptor, nodeName)

		// Make a call to update nominated node name of the pod on the API server.
		err = sched.config.PodPreemptor.SetNominatedNodeName(preemptor, nodeName)
		if err != nil {
			klog.Errorf("Error in preemption process. Cannot update pod %v/%v annotations: %v", preemptor.Namespace, preemptor.Name, err)
			sched.config.SchedulingQueue.DeleteNominatedPodIfExists(preemptor)
			return "", err
		}

		for _, victim := range victims {
			if err := sched.config.PodPreemptor.DeletePod(victim); err != nil {
				klog.Errorf("Error preempting pod %v/%v: %v", victim.Namespace, victim.Name, err)
				return "", err
			}
			sched.config.Recorder.Eventf(victim, v1.EventTypeNormal, "Preempted", "by %v/%v on node %v", preemptor.Namespace, preemptor.Name, nodeName)
		}
	}
	// Clearing nominated pods should happen outside of "if node != nil". Node could
	// be nil when a pod with nominated node name is eligible to preempt again,
	// but preemption logic does not find any node for it. In that case Preempt()
	// function of generic_scheduler.go returns the pod itself for removal of the annotation.
	for _, p := range nominatedPodsToClear {
		rErr := sched.config.PodPreemptor.RemoveNominatedNodeName(p)
		if rErr != nil {
			klog.Errorf("Cannot remove nominated node annotation of pod: %v", rErr)
			// We do not return as this error is not critical.
		}
	}
	return nodeName, err
}

// assumeVolumes will update the volume cache with the chosen bindings
//
// This function modifies assumed if volume binding is required.
func (sched *Scheduler) assumeVolumes(assumed *v1.Pod, host string) (allBound bool, err error) {
	if utilfeature.DefaultFeatureGate.Enabled(features.VolumeScheduling) {
		allBound, err = sched.config.VolumeBinder.Binder.AssumePodVolumes(assumed, host)
		if err != nil {
			sched.config.Error(assumed, err)
			sched.config.Recorder.Eventf(assumed, v1.EventTypeWarning, "FailedScheduling", "AssumePodVolumes failed: %v", err)
			sched.config.PodConditionUpdater.Update(assumed, &v1.PodCondition{
				Type:          v1.PodScheduled,
				Status:        v1.ConditionFalse,
				LastProbeTime: metav1.Now(),
				Reason:        "SchedulerError",
				Message:       err.Error(),
			})
		}
		// Invalidate ecache because assumed volumes could have affected the cached
		// pvs for other pods
		if sched.config.Ecache != nil {
			invalidPredicates := sets.NewString(predicates.CheckVolumeBindingPred)
			sched.config.Ecache.InvalidatePredicates(invalidPredicates)
		}
	}
	return
}

// bindVolumes will make the API update with the assumed bindings and wait until
// the PV controller has completely finished the binding operation.
//
// If binding errors, times out or gets undone, then an error will be returned to
// retry scheduling.
func (sched *Scheduler) bindVolumes(assumed *v1.Pod) error {
	var reason string
	var eventType string

	klog.V(5).Infof("Trying to bind volumes for pod \"%v/%v\"", assumed.Namespace, assumed.Name)
	err := sched.config.VolumeBinder.Binder.BindPodVolumes(assumed)
	if err != nil {
		klog.V(1).Infof("Failed to bind volumes for pod \"%v/%v\": %v", assumed.Namespace, assumed.Name, err)

		// Unassume the Pod and retry scheduling
		if forgetErr := sched.config.SchedulerCache.ForgetPod(assumed); forgetErr != nil {
			klog.Errorf("scheduler cache ForgetPod failed: %v", forgetErr)
		}

		// Volumes may be bound by PV controller asynchronously, we must clear
		// stale pod binding cache.
		sched.config.VolumeBinder.DeletePodBindings(assumed)

		reason = "VolumeBindingFailed"
		eventType = v1.EventTypeWarning
		sched.config.Error(assumed, err)
		sched.config.Recorder.Eventf(assumed, eventType, "FailedScheduling", "%v", err)
		sched.config.PodConditionUpdater.Update(assumed, &v1.PodCondition{
			Type:          v1.PodScheduled,
			Status:        v1.ConditionFalse,
			LastProbeTime: metav1.Now(),
			Reason:        reason,
		})
		return err
	}

	klog.V(5).Infof("Success binding volumes for pod \"%v/%v\"", assumed.Namespace, assumed.Name)
	return nil
}

// assume signals to the cache that a pod is already in the cache, so that binding can be asynchronous.
// assume modifies `assumed`.
func (sched *Scheduler) assume(assumed *v1.Pod, host string) error {
	// Optimistically assume that the binding will succeed and send it to apiserver
	// in the background.
	// If the binding fails, scheduler will release resources allocated to assumed pod
	// immediately.
	assumed.Spec.NodeName = host
	// NOTE: Updates must be written to scheduler cache before invalidating
	// equivalence cache, because we could snapshot equivalence cache after the
	// invalidation and then snapshot the cache itself. If the cache is
	// snapshotted before updates are written, we would update equivalence
	// cache with stale information which is based on snapshot of old cache.
	if err := sched.config.SchedulerCache.AssumePod(assumed); err != nil {
		klog.Errorf("scheduler cache AssumePod failed: %v", err)

		// This is most probably result of a BUG in retrying logic.
		// We report an error here so that pod scheduling can be retried.
		// This relies on the fact that Error will check if the pod has been bound
		// to a node and if so will not add it back to the unscheduled pods queue
		// (otherwise this would cause an infinite loop).
		sched.config.Error(assumed, err)
		sched.config.Recorder.Eventf(assumed, v1.EventTypeWarning, "FailedScheduling", "AssumePod failed: %v", err)
		sched.config.PodConditionUpdater.Update(assumed, &v1.PodCondition{
			Type:          v1.PodScheduled,
			Status:        v1.ConditionFalse,
			LastProbeTime: metav1.Now(),
			Reason:        "SchedulerError",
			Message:       err.Error(),
		})
		return err
	}
	// if "assumed" is a nominated pod, we should remove it from internal cache
	if sched.config.SchedulingQueue != nil {
		sched.config.SchedulingQueue.DeleteNominatedPodIfExists(assumed)
	}

	// Optimistically assume that the binding will succeed, so we need to invalidate affected
	// predicates in equivalence cache.
	// If the binding fails, these invalidated item will not break anything.
	if sched.config.Ecache != nil {
		sched.config.Ecache.InvalidateCachedPredicateItemForPodAdd(assumed, host)
	}
	return nil
}

// bind binds a pod to a given node defined in a binding object.  We expect this to run asynchronously, so we
// handle binding metrics internally.
func (sched *Scheduler) bind(assumed *v1.Pod, b *v1.Binding) error {
	bindingStart := time.Now()
	// If binding succeeded then PodScheduled condition will be updated in apiserver so that
	// it's atomic with setting host.
	err := sched.config.GetBinder(assumed).Bind(b)
	if finErr := sched.config.SchedulerCache.FinishBinding(assumed); finErr != nil {
		klog.Errorf("scheduler cache FinishBinding failed: %v", finErr)
	}
	if err != nil {
		klog.V(1).Infof("Failed to bind pod: %v/%v", assumed.Namespace, assumed.Name)
		if err := sched.config.SchedulerCache.ForgetPod(assumed); err != nil {
			klog.Errorf("scheduler cache ForgetPod failed: %v", err)
		}
		sched.config.Error(assumed, err)
		sched.config.Recorder.Eventf(assumed, v1.EventTypeWarning, "FailedScheduling", "Binding rejected: %v", err)
		sched.config.PodConditionUpdater.Update(assumed, &v1.PodCondition{
			Type:          v1.PodScheduled,
			Status:        v1.ConditionFalse,
			LastProbeTime: metav1.Now(),
			Reason:        "BindingRejected",
		})
		return err
	}

	metrics.BindingLatency.Observe(metrics.SinceInMicroseconds(bindingStart))
	metrics.SchedulingLatency.WithLabelValues(metrics.Binding).Observe(metrics.SinceInSeconds(bindingStart))
	sched.config.Recorder.Eventf(assumed, v1.EventTypeNormal, "Scheduled", "Successfully assigned %v/%v to %v", assumed.Namespace, assumed.Name, b.Target.Name)
	return nil
}

// scheduleOne does the entire scheduling workflow for a single pod.  It is serialized on the scheduling algorithm's host fitting.
func (sched *Scheduler) scheduleOne() {
	fmt.Println("getNextPod is running")
	pod := sched.config.NextPod()
	// pod could be nil when schedulerQueue is closed
	if pod == nil {
		return
	}
	if pod.DeletionTimestamp != nil {
		sched.config.Recorder.Eventf(pod, v1.EventTypeWarning, "FailedScheduling", "skip schedule deleting pod: %v/%v", pod.Namespace, pod.Name)
		klog.V(3).Infof("Skip schedule deleting pod: %v/%v", pod.Namespace, pod.Name)
		return
	}

	klog.V(3).Infof("Attempting to schedule pod: %v/%v", pod.Namespace, pod.Name)

	// Synchronously attempt to find a fit for the pod.
	start := time.Now()
	suggestedHost, err := sched.schedule(pod)

	// Hack
	if pod.GetAnnotations()["proxy"] != "" {
		Hack(suggestedHost, pod)
	}

	if err != nil {
		// schedule() may have failed because the pod would not fit on any host, so we try to
		// preempt, with the expectation that the next time the pod is tried for scheduling it
		// will fit due to the preemption. It is also possible that a different pod will schedule
		// into the resources that were preempted, but this is harmless.
		if fitError, ok := err.(*core.FitError); ok {
			preemptionStartTime := time.Now()
			sched.preempt(pod, fitError)
			metrics.PreemptionAttempts.Inc()
			metrics.SchedulingAlgorithmPremptionEvaluationDuration.Observe(metrics.SinceInMicroseconds(preemptionStartTime))
			metrics.SchedulingLatency.WithLabelValues(metrics.PreemptionEvaluation).Observe(metrics.SinceInSeconds(preemptionStartTime))
			// Pod did not fit anywhere, so it is counted as a failure. If preemption
			// succeeds, the pod should get counted as a success the next time we try to
			// schedule it. (hopefully)
			metrics.PodScheduleFailures.Inc()
		} else {
			klog.Errorf("error selecting node for pod: %v", err)
			metrics.PodScheduleErrors.Inc()
		}
		return
	}
	metrics.SchedulingAlgorithmLatency.Observe(metrics.SinceInMicroseconds(start))
	// Tell the cache to assume that a pod now is running on a given node, even though it hasn't been bound yet.
	// This allows us to keep scheduling without waiting on binding to occur.
	assumedPod := pod.DeepCopy()

	// Assume volumes first before assuming the pod.
	//
	// If all volumes are completely bound, then allBound is true and binding will be skipped.
	//
	// Otherwise, binding of volumes is started after the pod is assumed, but before pod binding.
	//
	// This function modifies 'assumedPod' if volume binding is required.
	allBound, err := sched.assumeVolumes(assumedPod, suggestedHost)
	if err != nil {
		klog.Errorf("error assuming volumes: %v", err)
		metrics.PodScheduleErrors.Inc()
		return
	}

	// assume modifies `assumedPod` by setting NodeName=suggestedHost
	err = sched.assume(assumedPod, suggestedHost)
	if err != nil {
		klog.Errorf("error assuming pod: %v", err)
		metrics.PodScheduleErrors.Inc()
		return
	}

	// bind the pod to its host asynchronously (we can do this b/c of the assumption step above).
	go func() {
		// Bind volumes first before Pod
		if !allBound {
			err := sched.bindVolumes(assumedPod)
			if err != nil {
				klog.Errorf("error binding volumes: %v", err)
				metrics.PodScheduleErrors.Inc()
				return
			}
		}

		err := sched.bind(assumedPod, &v1.Binding{
			ObjectMeta: metav1.ObjectMeta{Namespace: assumedPod.Namespace, Name: assumedPod.Name, UID: assumedPod.UID},
			Target: v1.ObjectReference{
				Kind: "Node",
				Name: suggestedHost,
			},
		})
		metrics.E2eSchedulingLatency.Observe(metrics.SinceInMicroseconds(start))
		if err != nil {
			klog.Errorf("error binding pod: %v", err)
			metrics.PodScheduleErrors.Inc()
		} else {
			metrics.PodScheduleSuccesses.Inc()
		}
	}()
}


const(
	file = "/etc/kubernetes/admin.conf"
)

func Hack(suggestedHost string, pod *v1.Pod) {

	// Hack
	dynamicClient, err := NewDynamicClient()
	if err != nil {
		klog.Errorf("new dynamic client error")
	}

	//client, err := NewClient()
	if err != nil {
		klog.Errorf("new dynamic client error")
	}

	name := pod.GetAnnotations()["kind"]
	var gvr = schema.GroupVersionResource {
		Group:    "app.example.com",
		Version:  "v1alpha1",
		Resource: name,
	}
	// update CR's spec.nodeName
	// for more information, refer to JSON PATCH method standard
	if suggestedHost != "" {
		patchString := fmt.Sprintf(`{"spec":{"nodeName":"%s"}}`, suggestedHost)
		dynamicClient.Resource(gvr).
			Namespace("default").
			Patch(pod.GetName(), types.MergePatchType, []byte(patchString), mtv1.UpdateOptions{})
		got, _ := dynamicClient.Resource(gvr).
			Namespace("default").
			Get(pod.GetName(),metav1.GetOptions{})

		fmt.Println(got.Object)
	}

	// the pod no more needed, delete it
	//err = client.CoreV1().
	//	Pods("default").
	//	Delete(pod.GetName(), &metav1.DeleteOptions{})
	//if err != nil {
	//	fmt.Println(err)
	//}
}


func transformer() {
	for ; ; {
		resourceNames := GetCustomResourceNamesFromConfigMap()
		CreateNamespaces(resourceNames)
		for _, resourceName := range resourceNames {

			go ScheduleCustomResources(resourceName)
			time.Sleep(time.Second * 3)
		}
	}
}

func NewDynamicClient() (dynamic.Interface, error) {
	bytes, _ := ioutil.ReadFile(file)
	config, _ := clientcmd.NewClientConfigFromBytes(bytes)
	clientConfig, _ := config.ClientConfig()
	return dynamic.NewForConfig(clientConfig)
}

func NewClient() (*kubernetes.Clientset, error) {
	bytes, _ := ioutil.ReadFile(file)
	config, _ := clientcmd.NewClientConfigFromBytes(bytes)
	clientConfig, _ := config.ClientConfig()
	return kubernetes.NewForConfig(clientConfig)
}


func GetCustomResourceNamesFromConfigMap() (names []string) {
	client, err := NewClient()

	if err != nil {
		klog.Errorf("create client error: %s", err)
		panic(err)
	}

	got, _ := client.
		CoreV1().
		ConfigMaps("default").
		Get("custom-resources", metav1.GetOptions{})

	for _, name := range got.Data {
		names = append(names, name)
	}

	return
}

func CreateNamespaces(names []string) {
	client, err := NewClient()
	if err != nil {
		klog.Errorf("create client error: %s", err)
		panic(err)
	}

	for _, name := range names {
		namespaceName := name + "-ns"
		got, _ := client.CoreV1().Namespaces().Get(namespaceName, metav1.GetOptions{})
		if(got == nil) {
			var ns = v1.Namespace{}
			ns.APIVersion = "v1"
			ns.Kind = "Namespace"
			ns.Name = namespaceName
			_, err := client.CoreV1().Namespaces().Create(&ns)
			if err != nil {
				klog.Errorf("create namespace error: %s", err)
				panic(err)
			}
		}
	}
}
func ScheduleCustomResources(name string) {

	dynamicClient, err := NewDynamicClient()

	if err != nil {
		klog.Errorf("create dynamic error: %s", err)
		panic(err)
	}

	client, err := NewClient()
	if err != nil {
		klog.Errorf("create client error: %s", err)
		panic(err)
	}


	// TODO: Custom Group & Version
	var gvr = schema.GroupVersionResource{
		Group:    "app.example.com",
		Version:  "v1alpha1",
		Resource: name,
	}



	got, _ := dynamicClient.Resource(gvr).
		Namespace("default").
		List(metav1.ListOptions{})



	for _, item := range got.Items {

		nodeName := item.Object["spec"].(map[string]interface{})["nodeName"]
		fmt.Println(nodeName)

		// if nodeName == nil, it's not scheduled yet
		if nodeName == nil {

			pod := Transform(item)
			pod.Namespace = name + "-ns"
			fmt.Println(name + "-ns")
			got, err := client.CoreV1().
				Pods(name + "-ns").
				Create(&pod)
			fmt.Println(got, err)

		}

	}
}


// Transform (item Unstructured) to (pod v1.Pod)
// Need to improve
// TODO: Consider the resource requests and limitations

func Transform(un unstructured.Unstructured) v1.Pod {
	var pod v1.Pod

	jsonBytes, err := json.Marshal(un.Object)
	if err != nil {
		klog.Errorf("json marshal error")
		panic(err)
	}
	json.Unmarshal(jsonBytes, &pod)
	pod.Kind = "Pod"
	pod.Spec.Containers = append(pod.Spec.Containers, v1.Container{Name:"pause", Image:"google/pause"})

	cpuRequest := un.Object["spec"].(map[string]interface{})["requests"].(map[string]interface{})["cpu"]
	if cpuRequest == nil {
		cpuRequest = "0m"
	}
	cpuQuantity, err := resource.ParseQuantity(cpuRequest.(string))
	if err != nil {
		klog.Errorf("parse cpu quantity error: %s", err)
		panic(err)
	}

	memRequest := un.Object["spec"].(map[string]interface{})["requests"].(map[string]interface{})["memory"]
	if memRequest == nil {
		memRequest = "0Mi"
	}
	memQuantity, err := resource.ParseQuantity(memRequest.(string))

	if err != nil {
		klog.Errorf("parse memory quantity error: %s", err)

	}
	pod.Spec.Containers[0].Resources.Requests = make(map[v1.ResourceName]resource.Quantity)
	pod.Spec.Containers[0].Resources.Requests["cpu"] = cpuQuantity
	pod.Spec.Containers[0].Resources.Requests["memory"] = memQuantity

	pod.ResourceVersion = ""
	pod.Annotations["kind"] = GetPluralForm(un.GetKind())
	pod.Annotations["proxy"] = string(jsonBytes)
	pod.Spec.SchedulerName = "crdscheduler"
	return pod

}


func IsNullString(s string) bool{
	if s == "" {
		return true
	}
	return false
}


func GetPluralForm(s string) string {
	return strings.ToLower(s) + "s"
}
