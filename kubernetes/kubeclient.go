package kubeclient

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	// Initialize all known client auth plugins.
	_ "k8s.io/client-go/plugin/pkg/client/auth/azure"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	_ "k8s.io/client-go/plugin/pkg/client/auth/openstack"
)

/*
Builds and returns ClientSet by using local KubeConfig
It also returns hostname since it is needed.
*/
func getLocalKubeClient() (*kubernetes.Clientset, string, error) {
	var kubeconfig *string
	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// fmt.Printf("Host: %v\n", config.Host)

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	return clientset, config.Host, nil
}

/*
This method installs namespace, cluster role binding and image pull secret
TODO: differenciate between already exists and other error types
*/
func SetupVampCredentials(clientset *kubernetes.Clientset, ns string, secretDataString string) error {
	nsSpec := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}}
	_, err_n := clientset.Core().Namespaces().Create(nsSpec)
	if err_n != nil {
		fmt.Printf("Warning: %v\n", err_n.Error())
	}
	// Create Cluster Role Binding Vamp Default Service Account
	clusterRoleBindingSpec := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{Name: ns + "-sa-cluster-admin-binding"},
		Subjects:   []rbacv1.Subject{rbacv1.Subject{Kind: "User", Name: "system:serviceaccount:" + ns + ":default", APIGroup: "rbac.authorization.k8s.io"}},
		RoleRef:    rbacv1.RoleRef{Kind: "ClusterRole", Name: "cluster-admin", APIGroup: "rbac.authorization.k8s.io"},
	}
	_, err_c := clientset.RbacV1().ClusterRoleBindings().Create(clusterRoleBindingSpec)
	if err_c != nil {
		fmt.Printf("Warning: %v\n", err_c.Error())
	}
	// Create Image Pull Secret
	pullSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "vamp2imagepull", Namespace: ns},
		Data: map[string][]byte{
			".dockercfg": []byte(secretDataString),
		},
		Type: "kubernetes.io/dockercfg",
	}
	_, err_s := clientset.Core().Secrets(ns).Create(pullSecret)
	if err_s != nil {
		fmt.Printf("Warning: %v\n", err_s.Error())
	}
	return nil
}

func BootstrapVampService() (string, string, string, error) {
	// create the clientset
	clientset, host, err := getLocalKubeClient()
	if err != nil {
		panic(err.Error())
	}
	ns := "vamp-system"
	secretDataString := "{\"https://index.docker.io/v1/\":{\"auth\":\"dmFtcDJwdWxsOnZhbXAycHVsbEZsdXg=\"}}"
	errSetup := SetupVampCredentials(clientset, ns, secretDataString)
	if errSetup != nil {
		fmt.Printf("Warning: %v\n", errSetup.Error())
		return host, "", "", errSetup
	}

	// This is end of setting up remote vamp set up
	// now we need to get information to connect to the cluster

	getOptions := metav1.GetOptions{}
	sa, err_sa := clientset.Core().ServiceAccounts(ns).Get("default", getOptions)
	if err_sa != nil {
		fmt.Printf("Warning: %v\n", err_sa.Error())
		return host, "", "", err_sa
	}

	saSecret, err_sa_secret := clientset.Core().Secrets(ns).Get(sa.Secrets[0].Name, getOptions)
	if err_sa_secret != nil {
		fmt.Printf("Warning: %v\n", err_sa_secret.Error())
		// This is a problem command should be re-tried by user
		return host, "", "", err_sa_secret
	}
	crt := string(saSecret.Data["ca.crt"])
	token := string(saSecret.Data["token"])

	return host, crt, token, nil
}

func InstallVampService() (string, string, string, error) {
	host, crt, token, errBootstap := BootstrapVampService()
	if errBootstap != nil {
		fmt.Printf("Warning: %v\n", errBootstap.Error())
		// This is a problem command should be re-tried by user
		return host, "", "", errBootstap
	}
	// Install Database or skip it
	// Deploy Db
	// Create service for db
	// Create internal service for sync
	// Create external service for vamp
	// Create Vamp Deployment
	// Wait for external service
	return host, crt, token, nil
}

func InstallMongoDB(clientset *kubernetes.Clientset, ns string) error {
	service := &apiv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: "vamp-mongodb",
		},
		Spec: apiv1.ServiceSpec{
			ClusterIP: "None",
			Selector: map[string]string{
				"app": "vamp-mongodb",
			},
			Ports: []apiv1.ServicePort{
				{
					Port:       27017,
					TargetPort: intstr.FromInt(27017),
				},
			},
		},
	}
	_, errService := clientset.Core().Services(ns).Create(service)
	if errService != nil {
		fmt.Printf("Warning: %v\n", errService.Error())
	}
	/*
	   apiVersion: apps/v1beta1
	   kind: StatefulSet
	   metadata:
	     namespace: vamp-system
	     name: mongo
	   spec:
	     serviceName: vamp-mongodb
	     replicas: 3
	     template:
	       metadata:
	         labels:
	           app: vamp-mongodb
	       spec:
	         terminationGracePeriodSeconds: 10
	         containers:
	           - name: mongo
	             image: mongo
	             command:
	               - mongod
	               - "--replSet"
	               - rs0
	               - "--bind_ip"
	               - 0.0.0.0
	               - "--smallfiles"
	               - "--noprealloc"
	             ports:
	               - containerPort: 27017
	             volumeMounts:
	               - name: mongo-persistent-storage
	                 mountPath: /data/db
	           - name: mongo-sidecar
	             image: cvallance/mongo-k8s-sidecar
	             env:
	               - name: MONGO_SIDECAR_POD_LABELS
	                 value: "app=vamp-mongodb"
	               - name: KUBERNETES_MONGO_SERVICE_NAME
	                 value: "vamp-mongodb"
	     volumeClaimTemplates:
	     - metadata:
	         name: mongo-persistent-storage
	         annotations:
	           volume.beta.kubernetes.io/storage-class: "standard"
	       spec:
	         accessModes: [ "ReadWriteOnce" ]
	         resources:
	           requests:
	             storage: 1Gi
	*/

	statefulSet := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name: "mongo",
		},
		Spec: appsv1.StatefulSetSpec{
			ServiceName: "vamp-mongodb",
			Replicas:    int32Ptr(3),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "vamp-mongodb",
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "vamp-mongodb",
					},
				},
				Spec: apiv1.PodSpec{
					TerminationGracePeriodSeconds: int64Ptr(10),
					Containers: []apiv1.Container{
						{
							Name:  "mongo",
							Image: "mongo",
							Command: []string{
								"mongod",
								"--replSet",
								"rs0",
								"--bind_ip",
								"0.0.0.0",
								"--smallfiles",
								"--noprealloc",
							},
							Ports: []apiv1.ContainerPort{
								{
									ContainerPort: 27017,
								},
							},
							VolumeMounts: []apiv1.VolumeMount{
								{
									Name:      "mongo-persistent-storage",
									MountPath: "/data/db",
								},
							},
						},
						{
							Name:  "mongo",
							Image: "mongo",
							Env: []apiv1.EnvVar{
								{
									Name:  "MONGO_SIDECAR_POD_LABELS",
									Value: "app=vamp-mongodb",
								},
								{
									Name:  "KUBERNETES_MONGO_SERVICE_NAME",
									Value: "vamp-mongodb",
								},
							},
						},
					},
				},
			},
			VolumeClaimTemplates: []apiv1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "mongo-persistent-storage",
						Annotations: map[string]string{
							"volume.beta.kubernetes.io/storage-class": "standard",
						},
					},
					Spec: apiv1.PersistentVolumeClaimSpec{
						AccessModes: []apiv1.PersistentVolumeAccessMode{
							"ReadWriteOnce",
						},
						Resources: apiv1.ResourceRequirements{
							Requests: apiv1.ResourceList{
								"storage": resource.MustParse("1Gi"),
							},
						},
					},
				},
			},
		},
	}
	_, errStatefulSet := clientset.AppsV1().StatefulSets(ns).Create(statefulSet)
	if errStatefulSet != nil {
		fmt.Printf("Warning: %v\n", errStatefulSet.Error())
	}
	return nil
}

/*
deployment := &appsv1.Deployment{
  ObjectMeta: metav1.ObjectMeta{
    Name: "demo-deployment",
  },
  Spec: appsv1.DeploymentSpec{
    Replicas: int32Ptr(2),
    Selector: &metav1.LabelSelector{
      MatchLabels: map[string]string{
        "app": "demo",
      },
    },
    Template: apiv1.PodTemplateSpec{
      ObjectMeta: metav1.ObjectMeta{
        Labels: map[string]string{
          "app": "demo",
        },
      },
      Spec: apiv1.PodSpec{
        Containers: []apiv1.Container{
          {
            Name:  "web",
            Image: "nginx:1.12",
            Ports: []apiv1.ContainerPort{
              {
                Name:          "http",
                Protocol:      apiv1.ProtocolTCP,
                ContainerPort: 80,
              },
            },
          },
        },
      },
    },
  },
}
fmt.Printf("Name: %v\n", deployment.ObjectMeta.Name)
_, errDeployment := clientset.AppsV1().Deployments(ns).Create(deployment)
if errDeployment != nil {
  fmt.Printf("Warning: %v\n", errDeployment.Error())
}
*/

func Run() {

	// create the clientset
	clientset, _, err := getLocalKubeClient()
	if err != nil {
		panic(err.Error())
	}
	for {
		pods, err := clientset.CoreV1().Pods("").List(metav1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}
		fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))

		// Examples for error handling:
		// - Use helper functions like e.g. errors.IsNotFound()
		// - And/or cast to StatusError and use its properties like e.g. ErrStatus.Message
		namespace := "vamp-demo"
		pod := "deployment1-7bfd78cf7c-4c2g2"
		_, err = clientset.CoreV1().Pods(namespace).Get(pod, metav1.GetOptions{})
		if errors.IsNotFound(err) {
			fmt.Printf("Pod %s in namespace %s not found\n", pod, namespace)
		} else if statusError, isStatus := err.(*errors.StatusError); isStatus {
			fmt.Printf("Error getting pod %s in namespace %s: %v\n",
				pod, namespace, statusError.ErrStatus.Message)
		} else if err != nil {
			panic(err.Error())
		} else {
			fmt.Printf("Found pod %s in namespace %s\n", pod, namespace)
		}

		time.Sleep(10 * time.Second)
	}
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

func int32Ptr(i int32) *int32 { return &i }
func int64Ptr(i int64) *int64 { return &i }
