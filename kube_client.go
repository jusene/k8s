package utils

import (
	"context"
	"fmt"
	config2 "gitlab.hho-inc.com/devops/oops-api/config"
	v1 "k8s.io/api/apps/v1"
	v1app "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type Kube interface {
	ListDeployment(app, proj, env string) (v1.Deployment, error)
	ListPods(app, proj, env string) ([]v1app.Pod, error)
	ListServices(app, proj, env string) (v1app.Service, error)
	ListNodes() ([]v1app.Node, error)
	ListNameSpaces() ([]v1app.Namespace, error)
	ListNameSpacePods(namespace string) ([]v1app.Pod, error)
	ListNameSpaceDeployments(namespace string) ([]v1.Deployment, error)
	ListNameSpaceServices(namespace string) ([]v1app.Service, error)

	CreateNameSpace(namespace string) (*v1app.Namespace, error)
	CreateDeployment(app, proj, env string, repl int32) (*v1.Deployment, error)
}

type KubeClient struct {
	clientSet *kubernetes.Clientset
}

func NewKubeClient(cluster string) *KubeClient {
	c, err := clientcmd.BuildConfigFromFlags("", "conf/"+config2.KubeConfigMap[cluster])
	if err != nil {
		panic(err)
	}

	clientSet, err := kubernetes.NewForConfig(c)
	if err != nil {
		panic(err)
	}

	return &KubeClient{
		clientSet,
	}
}

func (kube *KubeClient) ListDeployment(app, proj, env string) (v1.Deployment, error) {
	deploys, err := kube.clientSet.AppsV1().Deployments(fmt.Sprintf("%s-%s", proj, env)).
		List(context.TODO(), metav1.ListOptions{LabelSelector: fmt.Sprintf("name=%s", app), Limit: 100})
	if err != nil {
		return v1.Deployment{}, err
	}

	return deploys.Items[0], nil
}

func (kube *KubeClient) ListPods(app, proj, env string) ([]v1app.Pod, error) {
	pods, err := kube.clientSet.CoreV1().Pods(fmt.Sprintf("%s-%s", proj, env)).
		List(context.TODO(), metav1.ListOptions{LabelSelector: fmt.Sprintf("name=%s", app), Limit: 100})
	if err != nil {
		return []v1app.Pod{}, err
	}

	return pods.Items, nil
}

func (kube *KubeClient) ListServices(app, proj, env string) (v1app.Service, error) {
	services, err := kube.clientSet.CoreV1().Services(fmt.Sprintf("%s-%s", proj, env)).
		List(context.TODO(), metav1.ListOptions{LabelSelector: fmt.Sprintf("name=%s", app), Limit: 100})
	fmt.Println(services)
	if err != nil {
		return v1app.Service{}, err
	}

	return services.Items[0], nil
}

func (kube *KubeClient) ListNodes() ([]v1app.Node, error){
	nodes, err := kube.clientSet.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return []v1app.Node{}, err
	}

	return nodes.Items, nil
}

func (kube *KubeClient) ListNameSpaces() ([]v1app.Namespace, error) {
	namespace, err := kube.clientSet.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return []v1app.Namespace{}, err
	}

	return namespace.Items, nil
}

func (kube *KubeClient) ListNameSpacePods(namespace string) ([]v1app.Pod, error) {
	pods, err := kube.clientSet.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return []v1app.Pod{}, err
	}

	return pods.Items, nil
}

func (kube *KubeClient) ListNameSpaceDeployments(namespace string) ([]v1.Deployment, error) {
	deployments, err := kube.clientSet.AppsV1().Deployments(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return []v1.Deployment{}, err
	}

	return deployments.Items, nil
}

func (kube *KubeClient) ListNameSpaceServices(namespace string) ([]v1app.Service, error) {
	services, err := kube.clientSet.CoreV1().Services(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return []v1app.Service{}, err
	}

	return services.Items, nil
}

func (kube *KubeClient) CreateNameSpace(namespace string) (*v1app.Namespace, error) {
	result, err := kube.clientSet.CoreV1().Namespaces().Create(context.TODO(), &v1app.Namespace{
		TypeMeta:   metav1.TypeMeta{
			Kind:       "Namespace",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:                       namespace,
		},
	}, metav1.CreateOptions{})

	if err != nil {
		return &v1app.Namespace{}, err
	}

	return result, nil
}

func (kube *KubeClient) CreateDeployment(app, proj, env string, repl int32) (*v1.Deployment, error) {
	var labels = map[string]string{
		"name": app,
	}

	container := v1app.Container{
		Name:                     app,
		Image:                    fmt.Sprintf("reg.hho-inc.com/%s-%s/%s:%s", proj, env, app, "latest"), 
		Ports:                    []v1app.ContainerPort{
			v1app.ContainerPort{
				Name:          "app",
				ContainerPort: 7100,
			},
		},
		Env:                      []v1app.EnvVar{
			v1app.EnvVar{
				Name:      "TZ",
				Value:     "Asia/Shanghai",
			},
			v1app.EnvVar{
				Name:      "LANGUAGE",
				Value:     "en_US.UTF-8",
			},
			v1app.EnvVar{
				Name:      "LC_ALL",
				Value:     "en_US.UTF-8",
			},
			v1app.EnvVar{
				Name:      "LANG",
				Value:     "en_US.UTF-8",
			},
			v1app.EnvVar{
				Name:      "POD_APP",
				Value:     app,
			},
			v1app.EnvVar{
				Name:      "POD_IP",
				ValueFrom: &v1app.EnvVarSource{
					FieldRef:         &v1app.ObjectFieldSelector{
						APIVersion: "v1",
						FieldPath:  "status.podIP",
					},
				},
			},
			v1app.EnvVar{
				Name:      "POD_NAME",
				ValueFrom: &v1app.EnvVarSource{
					FieldRef:         &v1app.ObjectFieldSelector{
						APIVersion: "v1",
						FieldPath:  "metadata.name",
					},
				},
			},
			v1app.EnvVar{
				Name:      "POD_NAMESPACE",
				ValueFrom: &v1app.EnvVarSource{
					FieldRef:         &v1app.ObjectFieldSelector{
						APIVersion: "v1",
						FieldPath:  "metadata.namespace",
					},
				},
			},
		},
		Resources:                v1app.ResourceRequirements{
			Limits:   nil,
			Requests: nil,
		},
		VolumeMounts:             nil,
		VolumeDevices:            nil,
		LivenessProbe:            nil,
		ReadinessProbe:           nil,
		StartupProbe:             nil,
		Lifecycle:                nil,
		TerminationMessagePath:   "",
		TerminationMessagePolicy: "",
		ImagePullPolicy:          v1app.PullIfNotPresent,
	}

	result, err := kube.clientSet.AppsV1().Deployments(fmt.Sprintf("%s-%s", proj, env)).Create(context.TODO(), &v1.Deployment{
		TypeMeta:   metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:                       app,
			Namespace:                  fmt.Sprintf("%s-%s", proj, env),
			Labels: 					labels,
		},
		Spec:       v1.DeploymentSpec{
			Replicas:                &repl,
			Selector:                &metav1.LabelSelector{
				MatchLabels:            labels,
			},
			Template:                v1app.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:                       app,
					Namespace:                  fmt.Sprintf("%s-%s", proj, env),
					Labels:                     labels,
				},
				Spec:       v1app.PodSpec{
					Containers:                    []v1app.Container{container},
					RestartPolicy:                 "Always",
					TerminationGracePeriodSeconds: nil,
					ActiveDeadlineSeconds:         nil,
					DNSPolicy:                     "",
					NodeSelector:                  nil,
					ServiceAccountName:            "",
					DeprecatedServiceAccount:      "",
					AutomountServiceAccountToken:  func () *bool {var i = false; return &i}(),
					ShareProcessNamespace:         nil,
					ImagePullSecrets:              nil,
					Hostname:                      "",
					Subdomain:                     "",
					Affinity:                      nil,
					SchedulerName:                 "",
					Tolerations:                   nil,
					HostAliases:                   nil,
					PriorityClassName:             "",
					Priority:                      nil,
					DNSConfig:                     nil,
					ReadinessGates:                nil,
					RuntimeClassName:              nil,
					EnableServiceLinks:            nil,
					PreemptionPolicy:              nil,
					Overhead:                      nil,
					TopologySpreadConstraints:     nil,
					SetHostnameAsFQDN:             nil,
					OS:                            nil,
				},
			},
			Strategy:                v1.DeploymentStrategy{},
			MinReadySeconds:         0,
			RevisionHistoryLimit:    nil,
			Paused:                  false,
			ProgressDeadlineSeconds: nil,
		},
	}, metav1.CreateOptions{})

	if err != nil {
		return &v1.Deployment{}, err
	}

	return result, nil

}
