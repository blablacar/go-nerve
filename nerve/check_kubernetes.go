package nerve

import (
	"sync"

	"github.com/n0rad/go-erlog/data"
	"github.com/n0rad/go-erlog/errs"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"k8s.io/client-go/rest"
)

type CheckKubernetes struct {
	CheckCommon
	Namespace string
	PodName   string

	clientset *kubernetes.Clientset
}

func NewCheckKubernetes() *CheckKubernetes {
	return &CheckKubernetes{
		Namespace: "default",
	}
}

func (x *CheckKubernetes) Run(statusChange chan Check, stop <-chan struct{}, doneWait *sync.WaitGroup) {
	x.CommonRun(x, statusChange, stop, doneWait)
}

func (x *CheckKubernetes) Init(s *Service) error {
	if err := x.CheckCommon.CommonInit(s); err != nil {
		return err
	}
	if x.PodName == "" {
		return errs.With("podName is mandatory")
	}
	if x.Namespace == "" {
		return errs.With("namespace is mandatory")
	}
	x.fields = x.fields.WithField("podName", x.PodName).WithField("namespace", x.Namespace)

	config, err := rest.InClusterConfig()
	if err != nil {
		return errs.WithEF(err, x.fields, "Fail to create in cluster config")
	}
	clientset, err := kubernetes.NewForConfig(config)
	x.clientset = clientset
	if err != nil {
		return errs.WithEF(err, x.fields, "Fail to create client from config")
	}
	return nil
}

func (x *CheckKubernetes) Check() error {
	pod, err := x.clientset.CoreV1().Pods(x.Namespace).Get(x.PodName, metav1.GetOptions{})
	if err != nil {
		return errs.WithEF(err, x.fields, "Fail to get pod")
	}

	ready := getConditionReadyFromStatus(pod.Status)
	if ready == nil {
		return errs.With("Failed to get pod condition ready")
	}

	if ready.Status != v1.ConditionTrue {
		return errs.WithF(data.WithField("status", ready.Status).
			WithField("reason", ready.Reason).
			WithField("message", ready.Message), "Pod is not ready")
	}
	return nil
}

func getConditionReadyFromStatus(status v1.PodStatus) *v1.PodCondition {
	for i := range status.Conditions {
		if status.Conditions[i].Type == v1.PodReady {
			return &status.Conditions[i]
		}
	}
	return nil
}
