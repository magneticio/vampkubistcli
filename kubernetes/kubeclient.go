package kubeclient

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"

	"github.com/cenkalti/backoff"
	"github.com/magneticio/vampkubistcli/client"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/cert"
	"k8s.io/client-go/util/retry"
	// Initialize all known client auth plugins.
	_ "k8s.io/client-go/plugin/pkg/client/auth/azure"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	_ "k8s.io/client-go/plugin/pkg/client/auth/openstack"
)

type VampConfig struct {
	RootPassword string `yaml:"rootPassword,omitempty" json:"rootPassword,omitempty"`
	DatabaseUrl  string `yaml:"databaseUrl,omitempty" json:"databaseUrl,omitempty"`
	RepoUsername string `yaml:"repoUsername,omitempty" json:"repoUsername,omitempty"`
	RepoPassword string `yaml:"repoPassword,omitempty" json:"repoPassword,omitempty"`
	VampVersion  string `yaml:"vampVersion,omitempty" json:"vampVersion,omitempty"`
}

func GetKubeConfigPath(configPath string) *string {
	if configPath == "" {
		home := homeDir()
		path := filepath.Join(home, ".kube", "config")
		return &path
	}
	return &configPath
}

/*
Builds and returns ClientSet by using local KubeConfig
It also returns hostname since it is needed.
*/
func getLocalKubeClient() (*kubernetes.Clientset, string, error) {
	kubeconfig := GetKubeConfigPath("")
	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}
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
	secretErr := CreateOrUpdateSecret(clientset, ns, pullSecret)
	if secretErr != nil {
		fmt.Printf("Warning: %v\n", secretErr.Error())
		return secretErr
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

func InstallVampService(config *VampConfig) (string, []byte, []byte, error) {
	host, _, _, errBootstap := BootstrapVampService()
	if errBootstap != nil {
		fmt.Printf("Warning: %v\n", errBootstap.Error())
		// This is a problem command should be re-tried by user
		return host, nil, nil, errBootstap
	}
	// create the clientset
	clientset, host, err := getLocalKubeClient()
	if err != nil {
		panic(err.Error())
	}
	ns := "vamp-system"
	// Install Database or skip it
	// Deploy Db
	if config.DatabaseUrl == "" {
		installMongoErr := InstallMongoDB(clientset, ns)
		if installMongoErr != nil {
			return "", nil, nil, installMongoErr
		}
		config.DatabaseUrl = "mongodb://mongo-0.vamp-mongodb:27017,mongo-1.vamp-mongodb:27017,mongo-2.vamp-mongodb:27017"
	}
	// Deploy vamp
	url, cert, key, installVampErr := InstallVamp(clientset, ns, config)
	if installVampErr != nil {
		return "", nil, nil, installVampErr
	}
	// NewRestClient(url string, token string, isDebug bool, cert string)
	/* ip, getIpError := GetServiceExternalIP(clientset, ns, "vamp")
	if getIpError != nil {
		return "", "", "", getIpError
	} */
	// TODO: add certificates && HTTPS
	// url := "http://" + ip + ":8888"
	// Wait for external service
	CheckAndWaitForService(*url, cert)
	return *url, cert, key, nil
}

func CheckAndWaitForService(url string, cert []byte) error {
	vampClient := client.NewRestClient(url, "", false, string(cert))
	count := 1
	operation := func() error {
		fmt.Printf("Pinging the service trial: %v\n", count)
		count += 1
		pong, pingErr := vampClient.Ping()
		if pingErr != nil {
			fmt.Printf("Failed to ping the service: %v\n", pingErr)
			return pingErr
		}
		if pong {
			fmt.Printf("Connection is available\n")
			return nil
		}
		return errors.New("Service is not available yet.")
	}

	err := backoff.Retry(operation, backoff.NewExponentialBackOff())
	if err != nil {
		return err
	}
	return nil
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
	errService := CreateOrUpdateService(clientset, ns, service)
	if errService != nil {
		fmt.Printf("Warning: %v\n", errService.Error())
		return errService
	}
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
							Name:  "mongo-sidecar",
							Image: "cvallance/mongo-k8s-sidecar",
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
	errStatefulSet := CreateOrUpdateStatefulSet(clientset, ns, statefulSet)
	if errStatefulSet != nil {
		fmt.Printf("Warning: %v\n", errStatefulSet.Error())
		return errStatefulSet
	}
	return nil
}

func InstallVamp(clientset *kubernetes.Clientset, ns string, config *VampConfig) (*string, []byte, []byte, error) {
	// Create Image Pull Secret
	dockerRepoAuth := base64.StdEncoding.EncodeToString([]byte(config.RepoUsername + ":" + config.RepoPassword))
	pullSecretDataString := "{\"https://index.docker.io/v1/\":{\"auth\":\"" + dockerRepoAuth + "\"}}"
	pullSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "vampkubistimagepull", Namespace: ns},
		Data: map[string][]byte{
			".dockercfg": []byte(pullSecretDataString),
		},
		Type: "kubernetes.io/dockercfg",
	}
	secretErr := CreateOrUpdateSecret(clientset, ns, pullSecret)
	if secretErr != nil {
		fmt.Printf("Warning: %v\n", secretErr.Error())
		return nil, nil, nil, secretErr
	}

	hazelcastService := &apiv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: "vamp-hazelcast",
		},
		Spec: apiv1.ServiceSpec{
			Selector: map[string]string{
				"app": "vamp",
			},
			Ports: []apiv1.ServicePort{
				{
					Protocol:   "TCP",
					Port:       5701,
					TargetPort: intstr.FromInt(5701),
				},
			},
		},
	}
	errHazelcastService := CreateOrUpdateService(clientset, ns, hazelcastService)
	if errHazelcastService != nil {
		fmt.Printf("Warning: %v\n", errHazelcastService.Error())
		return nil, nil, nil, errHazelcastService
	}
	vampService := &apiv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: "vamp",
		},
		Spec: apiv1.ServiceSpec{
			Type: "LoadBalancer",
			Selector: map[string]string{
				"app": "vamp",
			},
			Ports: []apiv1.ServicePort{
				{
					Protocol:   "TCP",
					Port:       8888,
					TargetPort: intstr.FromInt(8888),
				},
			},
		},
	}
	errVampService := CreateOrUpdateService(clientset, ns, vampService)
	if errVampService != nil {
		fmt.Printf("Warning: %v\n", errVampService.Error())
		return nil, nil, nil, errVampService
	}
	ip, getIpError := GetServiceExternalIP(clientset, ns, vampService.GetObjectMeta().GetName())
	if getIpError != nil {
		return nil, nil, nil, getIpError
	}
	// certificates
	cert, key, certError := cert.GenerateSelfSignedCertKey(ip, []net.IP{}, []string{})
	if certError != nil {
		return nil, nil, nil, certError
	}
	certSecretName := "certificates-for-" + ip

	// certSecret, getCertSecretErr := GetOpaqueSecret(clientset, ns, certSecretName)
	certSecretError := CreateOrUpdateOpaqueSecret(clientset, ns, certSecretName,
		map[string][]byte{
			"cert": cert,
			"key":  key,
		})
	if certSecretError != nil {
		fmt.Printf("Warning: %v\n", certSecretError.Error())
		return nil, nil, nil, certSecretError
	}
	// Create Root Password Secret
	paswordSecretErr := CreateOrUpdateOpaqueSecret(clientset, ns, "vamprootpassword",
		map[string][]byte{
			"password": []byte(config.RootPassword),
		})
	if paswordSecretErr != nil {
		fmt.Printf("Warning: %v\n", paswordSecretErr.Error())
		return nil, nil, nil, paswordSecretErr
	}
	vampDeployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "vamp",
			Labels: map[string]string{
				"app":        "vamp",
				"deployment": "vamp",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(3),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":        "vamp",
					"deployment": "vamp",
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":        "vamp",
						"deployment": "vamp",
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  "vamp",
							Image: "magneticio/vamp2:" + config.VampVersion,
							Ports: []apiv1.ContainerPort{
								{
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: 8888,
								},
								{
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: 5701,
								},
							},
							Env: []apiv1.EnvVar{
								{
									Name:  "MODE",
									Value: "IN_CLUSTER",
								},
								{
									Name:  "DBURL",
									Value: config.DatabaseUrl,
								},
								{
									Name:  "DBNAME",
									Value: "vamp",
								},
								{
									Name:  "API_SSL",
									Value: "enabled",
								},
								{
									Name: "API_PRIVATE_KEY",
									ValueFrom: &apiv1.EnvVarSource{
										SecretKeyRef: &apiv1.SecretKeySelector{
											LocalObjectReference: apiv1.LocalObjectReference{
												Name: certSecretName,
											},
											Key: "key",
										},
									},
								},
								{
									Name: "API_SERVER_CERTIFICATE",
									ValueFrom: &apiv1.EnvVarSource{
										SecretKeyRef: &apiv1.SecretKeySelector{
											LocalObjectReference: apiv1.LocalObjectReference{
												Name: certSecretName,
											},
											Key: "cert",
										},
									},
								},
								{
									Name: "ROOT_PASSWORD",
									ValueFrom: &apiv1.EnvVarSource{
										SecretKeyRef: &apiv1.SecretKeySelector{
											LocalObjectReference: apiv1.LocalObjectReference{
												Name: "vamprootpassword",
											},
											Key: "password",
										},
									},
								},
							},
						},
					},
					ImagePullSecrets: []apiv1.LocalObjectReference{
						{
							Name: "vampkubistimagepull",
						},
					},
				},
			},
		},
	}
	errDeployment := CreateOrUpdateDeployment(clientset, ns, vampDeployment)
	if errDeployment != nil {
		fmt.Printf("Warning: %v\n", errDeployment.Error())
		return nil, nil, nil, errDeployment
	}
	url := "https://" + ip + ":8888"
	return &url, cert, key, nil
}

func CreateOrUpdateDeployment(clientset *kubernetes.Clientset, ns string, deployment *appsv1.Deployment) error {
	fmt.Printf("CreateOrUpdateDeployment: %v\n", deployment.GetObjectMeta().GetName())
	deploymentsClient := clientset.AppsV1().Deployments(ns)
	_, errDeployment := deploymentsClient.Create(deployment)
	if errDeployment != nil {
		fmt.Printf("Warning: %v\n", errDeployment.Error())
		retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			// Retrieve the latest version of Deployment before attempting update
			// RetryOnConflict uses exponential backoff to avoid exhausting the apiserver
			_, getErr := deploymentsClient.Get(deployment.GetObjectMeta().GetName(), metav1.GetOptions{})
			if getErr != nil {
				panic(fmt.Errorf("Failed to get latest version of Deployment: %v", getErr))
			}
			_, updateErr := deploymentsClient.Update(deployment)
			return updateErr
		})
		if retryErr != nil {
			panic(fmt.Errorf("Update failed: %v", retryErr))
		}
	}
	return nil
}

func CreateOrUpdateService(clientset *kubernetes.Clientset, ns string, service *apiv1.Service) error {
	fmt.Printf("CreateOrUpdateService: %v\n", service.GetObjectMeta().GetName())
	servicesClient := clientset.Core().Services(ns)
	_, errService := servicesClient.Create(service)
	if errService != nil {
		fmt.Printf("Warning: %v\n", errService.Error())
		retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			// Retrieve the latest version of Service before attempting update
			// RetryOnConflict uses exponential backoff to avoid exhausting the apiserver
			currentService, getErr := servicesClient.Get(service.GetObjectMeta().GetName(), metav1.GetOptions{})
			if getErr != nil {
				panic(fmt.Errorf("Failed to get latest version of Service: %v", getErr))
			}
			service.Spec.ClusterIP = currentService.Spec.ClusterIP
			// TODO: increment resource version
			service.ObjectMeta.ResourceVersion = currentService.ObjectMeta.ResourceVersion
			_, updateErr := servicesClient.Update(service)
			return updateErr
		})
		if retryErr != nil {
			panic(fmt.Errorf("Update failed: %v", retryErr))
		}
	}
	return nil
}

func GetServiceExternalIP(clientset *kubernetes.Clientset, ns string, name string) (string, error) {
	fmt.Printf("GetServiceExternalIP: %v\n", name)
	servicesClient := clientset.Core().Services(ns)
	count := 1
	ip := ""
	operation := func() error {
		fmt.Printf("Getting External IP Trial: %v\n", count)
		count += 1
		currentService, getErr := servicesClient.Get(name, metav1.GetOptions{})
		if getErr != nil {
			panic(fmt.Errorf("Failed to get latest version of Service: %v", getErr))
		}
		ingress := currentService.Status.LoadBalancer.Ingress
		if ingress != nil && len(ingress) > 0 {
			ip = ingress[0].IP
			if ip != "" {
				return nil
			}
		}
		return errors.New("IP is not available yet.")
	}

	err := backoff.Retry(operation, backoff.NewExponentialBackOff())
	if err != nil {
		return "", err
	}

	return ip, nil
}

func CreateOrUpdateSecret(clientset *kubernetes.Clientset, ns string, secret *apiv1.Secret) error {
	fmt.Printf("CreateOrUpdateSecret: %v\n", secret.GetObjectMeta().GetName())
	secretsClient := clientset.Core().Secrets(ns)
	_, err := secretsClient.Create(secret)
	if err != nil {
		fmt.Printf("Warning: %v\n", err.Error())
		retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			// Retrieve the latest version of Secret before attempting update
			// RetryOnConflict uses exponential backoff to avoid exhausting the apiserver
			_, getErr := secretsClient.Get(secret.GetObjectMeta().GetName(), metav1.GetOptions{})
			if getErr != nil {
				panic(fmt.Errorf("Failed to get latest version of Secret: %v", getErr))
			}
			_, updateErr := secretsClient.Update(secret)
			return updateErr
		})
		if retryErr != nil {
			panic(fmt.Errorf("Update failed: %v", retryErr))
		}
	}
	return nil
}

func CreateOrUpdateStatefulSet(clientset *kubernetes.Clientset, ns string, statefulSet *appsv1.StatefulSet) error {
	fmt.Printf("CreateOrUpdateStatefulSet: %v\n", statefulSet.GetObjectMeta().GetName())
	statefulSetsClient := clientset.AppsV1().StatefulSets(ns)
	_, err := statefulSetsClient.Create(statefulSet)
	if err != nil {
		fmt.Printf("Warning: %v\n", err.Error())
		retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			// Retrieve the latest version of StatefulSet before attempting update
			// RetryOnConflict uses exponential backoff to avoid exhausting the apiserver
			_, getErr := statefulSetsClient.Get(statefulSet.GetObjectMeta().GetName(), metav1.GetOptions{})
			if getErr != nil {
				panic(fmt.Errorf("Failed to get latest version of StatefulSet: %v", getErr))
			}
			_, updateErr := statefulSetsClient.Update(statefulSet)
			return updateErr
		})
		if retryErr != nil {
			panic(fmt.Errorf("Update failed: %v", retryErr))
		}
	}
	return nil
}

func CreateOrUpdateOpaqueSecret(clientset *kubernetes.Clientset, ns string, name string, data map[string][]byte) error {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Data:       data,
		Type:       "Opaque",
	}
	return CreateOrUpdateSecret(clientset, ns, secret)
}

func GetOpaqueSecret(clientset *kubernetes.Clientset, ns string, name string) (map[string][]byte, error) {
	secretsClient := clientset.Core().Secrets(ns)
	secret, getError := secretsClient.Get(name, metav1.GetOptions{})
	if getError != nil {
		return nil, getError
	}
	return secret.Data, nil
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

func int32Ptr(i int32) *int32 { return &i }
func int64Ptr(i int64) *int64 { return &i }
