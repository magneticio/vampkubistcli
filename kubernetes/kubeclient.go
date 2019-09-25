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
	"github.com/magneticio/vampkubistcli/logging"
	"github.com/magneticio/vampkubistcli/models"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	apiv1 "k8s.io/api/core/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	apiutil "k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/cert"
	"k8s.io/client-go/util/retry"

	// Initialize all known client auth plugins.
	_ "k8s.io/client-go/plugin/pkg/client/auth/azure"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	_ "k8s.io/client-go/plugin/pkg/client/auth/openstack"
)

// Golang does't support struct constants
// Default values for an installation config
var DefaultVampConfig = models.VampConfig{
	DatabaseName:          "vamp",
	ImageName:             "magneticio/vampkubist",
	ImageTag:              "0.7.8",
	Mode:                  "IN_CLUSTER",
	AccessTokenExpiration: "10m",
	IstioAdapterImage:     "magneticio/vampkubist-istio-adapter-dev:latest",
	IstioInstallerImage:   "magneticio/vampistioinstaller:0.1.12",
}

// This is shared between installation and credentials, it is currently not configurable
// TODO: add it to VampConfig when it is configurable
const InstallationNamespace = "vamp-system"

var IsKubeClientInCluster = false

// VampClusterRoleBindingName contains name of ClusterRoleBinding for Vamp
const VampClusterRoleBindingName = InstallationNamespace + "-sa-cluster-admin-binding"

func VampConfigValidateAndSetupDefaults(config *models.VampConfig) (*models.VampConfig, error) {
	if config.RootPassword == "" {
		// This is enforced
		return config, errors.New("Root Password can not be empty.")
	}
	if config.RepoUsername == "" {
		return config, errors.New("Repo Username can not be empty.")
	}
	if config.RepoPassword == "" {
		return config, errors.New("Repo Password can not be empty.")
	}
	if config.DatabaseName == "" {
		config.DatabaseName = DefaultVampConfig.DatabaseName
		fmt.Printf("Database Name set to default value: %v\n", config.DatabaseName)
	}
	if config.ImageName == "" {
		config.ImageName = DefaultVampConfig.ImageName
		fmt.Printf("Image Name set to default value: %v\n", config.ImageName)
	}
	if config.ImageTag == "" {
		config.ImageTag = DefaultVampConfig.ImageTag
		fmt.Printf("Image Tag set to default value: %v\n", config.ImageTag)
	}
	if config.Mode != "IN_CLUSTER" &&
		config.Mode != "OUT_CLUSTER" &&
		config.Mode != "OUT_OF_CLUSTER" {
		config.Mode = DefaultVampConfig.Mode
		fmt.Printf("Vamp Mode set to default value: %v\n", config.Mode)
	}
	if config.AccessTokenExpiration == "" {
		config.AccessTokenExpiration = DefaultVampConfig.AccessTokenExpiration
		fmt.Printf("Access token expiration set to default value: %v\n", config.AccessTokenExpiration)
	}
	if config.IstioAdapterImage == "" {
		config.IstioAdapterImage = DefaultVampConfig.IstioAdapterImage
		fmt.Printf("Istio Adapter Image set to default value: %v\n", config.IstioAdapterImage)
	}
	if config.IstioInstallerImage == "" {
		config.IstioInstallerImage = DefaultVampConfig.IstioInstallerImage
		fmt.Printf("Istio Installer Image set to default value: %v\n", config.IstioInstallerImage)
	}
	if config.MaxReplicas == 0 {
		config.MaxReplicas = 6
		fmt.Printf("MaxReplicas set to %v\n", config.MaxReplicas)
	}
	return config, nil
}

/*
Tries to detect kubeconfig path if it is not explicitly set
*/
func GetKubeConfigPath(configPath string) *string {
	if configPath == "" {
		home := homeDir()
		path := filepath.Join(home, ".kube", "config")
		logging.Info("Using kube config path: %v\n", path)
		return &path
	}
	logging.Info("Using kube config path: %v\n", configPath)
	return &configPath
}

/*
Builds and returns ClientSet by using local KubeConfig
It also returns hostname since it is needed.
*/
func getLocalKubeClient(configPath string) (*kubernetes.Clientset, string, error) {
	if IsKubeClientInCluster {
		// creates the in-cluster config
		config, err := rest.InClusterConfig()
		if err != nil {
			return nil, "", errors.New(fmt.Sprintf("Kube Client can not be created due to %v", err.Error()))
		}
		// create the clientset
		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			return nil, "", err
		}
		return clientset, config.Host, nil
	}
	kubeconfigpath := GetKubeConfigPath(configPath)
	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfigpath)
	if err != nil {
		return nil, "", errors.New(fmt.Sprintf("Kube Client can not be created due to %v", err.Error()))
	}
	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, "", err
	}
	return clientset, config.Host, nil
}

/*
This method installs namespace, cluster role binding and image pull secret
TODO: differenciate between already exists and other error types
*/
func SetupVampCredentials(clientset *kubernetes.Clientset, ns string, rbName string) error {
	nsSpec := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}}
	_, namespaceCreationError := clientset.CoreV1().Namespaces().Create(nsSpec)
	if namespaceCreationError != nil {
		// TODO: handle already exists
		fmt.Printf("Warning: %v\n", namespaceCreationError.Error())
	}
	// Create Cluster Role Binding Vamp Default Service Account

	clusterRoleBindingSpec := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{Name: rbName},
		Subjects:   []rbacv1.Subject{rbacv1.Subject{Kind: "User", Name: "system:serviceaccount:" + ns + ":default", APIGroup: "rbac.authorization.k8s.io"}},
		RoleRef:    rbacv1.RoleRef{Kind: "ClusterRole", Name: "cluster-admin", APIGroup: "rbac.authorization.k8s.io"},
	}
	_, roleBindingCreationError := clientset.RbacV1().ClusterRoleBindings().Create(clusterRoleBindingSpec)
	if roleBindingCreationError != nil {
		// TODO: handle already exists
		fmt.Printf("Warning: %v\n", roleBindingCreationError.Error())
	}
	// Note: Image pull secret creation removed to only deploy public images on remote clusters
	return nil
}

func RemoveVampCredentials(clientset *kubernetes.Clientset, ns string, rbName string) error {
	if err := clientset.CoreV1().Namespaces().Delete(ns, nil); err != nil {
		fmt.Printf("Canot delete Vamp namespace %v - %v", ns, err)
		return err
	}
	if err := clientset.RbacV1().ClusterRoleBindings().Delete(rbName, nil); err != nil {
		fmt.Printf("Canot delete role bindings %v - %v", ns, err)
		return err
	}
	return nil
}

func BootstrapVampService(configPath string) (string, string, string, error) {
	// create the clientset
	clientset, host, err := getLocalKubeClient(configPath)
	if err != nil {
		// panic(err.Error())
		return "", "", "", err
	}
	ns := InstallationNamespace
	errSetup := SetupVampCredentials(clientset, ns, VampClusterRoleBindingName)
	if errSetup != nil {
		fmt.Printf("Warning: %v\n", errSetup.Error())
		return host, "", "", errSetup
	}

	// This is end of setting up remote vamp set up
	// now we need to get information to connect to the cluster

	getOptions := metav1.GetOptions{}
	sa, err_sa := clientset.CoreV1().ServiceAccounts(ns).Get("default", getOptions)
	if err_sa != nil {
		fmt.Printf("Warning: %v\n", err_sa.Error())
		return host, "", "", err_sa
	}

	saSecret, err_sa_secret := clientset.CoreV1().Secrets(ns).Get(sa.Secrets[0].Name, getOptions)
	if err_sa_secret != nil {
		fmt.Printf("Warning: %v\n", err_sa_secret.Error())
		// This is a problem command should be re-tried by user
		return host, "", "", err_sa_secret
	}
	crt := string(saSecret.Data["ca.crt"])
	token := string(saSecret.Data["token"])

	return host, crt, token, nil
}

func InstallVampService(config *models.VampConfig, configPath string) (string, []byte, []byte, error) {
	host, _, _, errBootstap := BootstrapVampService(configPath)
	if errBootstap != nil {
		fmt.Printf("Warning: %v\n", errBootstap.Error())
		// This is a problem command should be re-tried by user
		return host, nil, nil, errBootstap
	}
	// create the clientset
	clientset, host, err := getLocalKubeClient(configPath)
	if err != nil {
		// panic(err.Error())
		return "", nil, nil, err
	}
	ns := InstallationNamespace
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
	// this waits until service is accessible and cerficate is valid
	CheckAndWaitForService(*url, cert)
	return *url, cert, key, nil
}

func UninstallVampService(configPath string) error {
	clientset, _, err := getLocalKubeClient(configPath)
	if err != nil {
		return err
	}
	return RemoveVampCredentials(clientset, InstallationNamespace, VampClusterRoleBindingName)
}

func CheckAndWaitForService(url string, cert []byte) error {
	vampClient := client.NewRestClient(url, "", "", false, string(cert), nil)
	count := 1
	operation := func() error {
		fmt.Printf("Pinging the service trial %v\n", count)
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
							Image: "mongo:4.0",
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

func InstallVamp(clientset *kubernetes.Clientset, ns string, config *models.VampConfig) (*string, []byte, []byte, error) {
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

	certSecretName := "certificates-for-" + ip
	crt := []byte{}
	key := []byte{}
	certSecret, getCertSecretErr := GetOpaqueSecret(clientset, ns, certSecretName)
	if getCertSecretErr != nil {
		fmt.Printf("Warning: %v\n", getCertSecretErr.Error())
		crtGen, keyGen, certError := cert.GenerateSelfSignedCertKey(ip, []net.IP{}, []string{})
		if certError != nil {
			return nil, nil, nil, certError
		}
		crt = crtGen
		key = keyGen
		certSecretError := CreateOrUpdateOpaqueSecret(clientset, ns, certSecretName,
			map[string][]byte{
				"cert": crt,
				"key":  key,
			})
		if certSecretError != nil {
			fmt.Printf("Warning: %v\n", certSecretError.Error())
			return nil, nil, nil, certSecretError
		}
	} else {
		crt = certSecret["cert"]
		key = certSecret["key"]
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

	maxSurge := apiutil.FromInt(1)
	maxUnavailable := apiutil.FromInt(0)

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
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RollingUpdateDeploymentStrategyType,
				RollingUpdate: &appsv1.RollingUpdateDeployment{
					MaxSurge:       &maxSurge,
					MaxUnavailable: &maxUnavailable,
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
							Image: config.ImageName + ":" + config.ImageTag,
							Lifecycle: &corev1.Lifecycle{
								PreStop: &corev1.Handler{
									Exec: &corev1.ExecAction{
										Command: []string{"/bin/bash", "-c", "sleep 20"},
									},
								},
							},
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
							ReadinessProbe: &apiv1.Probe{
								Handler: apiv1.Handler{
									HTTPGet: &apiv1.HTTPGetAction{
										Path: "/ready",
										Port: apiutil.FromInt(8889),
									},
								},
								InitialDelaySeconds: 10,
								PeriodSeconds:       6,
								SuccessThreshold:    2,
								TimeoutSeconds:      2,
								FailureThreshold:    1,
							},
							LivenessProbe: &apiv1.Probe{
								Handler: apiv1.Handler{
									HTTPGet: &apiv1.HTTPGetAction{
										Path: "/",
										Port: apiutil.FromInt(8889),
									},
								},
								InitialDelaySeconds: 90,
								PeriodSeconds:       5,
								TimeoutSeconds:      2,
							},
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{"cpu": *resource.NewScaledQuantity(100, resource.Milli)},
							},
							Env: []apiv1.EnvVar{
								{
									Name:  "MODE",
									Value: config.Mode,
								},
								{
									Name:  "DBURL",
									Value: config.DatabaseUrl,
								},
								{
									Name:  "DBNAME",
									Value: config.DatabaseName,
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
									Name:  "API_EXTERNAL_HOST",
									Value: ip + ":8888",
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
								{
									Name:  "OAUTH_TOKEN_EXPIRATION",
									Value: config.AccessTokenExpiration,
								},
								{
									Name:  "ISTIO_INSTALLER_IMAGE",
									Value: config.IstioInstallerImage,
								},
								{
									Name:  "ISTIO_ADAPTER_IMAGE",
									Value: config.IstioAdapterImage,
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
		fmt.Printf("Warning: error during deployment - %v\n", errDeployment.Error())
		return nil, nil, nil, errDeployment
	}

	errHPA := CreateOrUpdateHPA(clientset, config)
	if errHPA != nil {
		fmt.Printf("Warning: error during hpa creation - %v\n", errHPA.Error())
		return nil, nil, nil, errHPA
	}

	url := "https://" + ip + ":8888"
	return &url, crt, key, nil
}

func CreateOrUpdateHPA(clientset *kubernetes.Clientset, config *models.VampConfig) error {
	hpa := &autoscalingv1.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name: "vamp",
			Labels: map[string]string{
				"app":        "vamp",
				"deployment": "vamp",
			},
		},
		Spec: autoscalingv1.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: autoscalingv1.CrossVersionObjectReference{
				Kind:       "Deployment",
				Name:       "vamp",
				APIVersion: "extensions/v1beta1",
			},
			MinReplicas:                    config.MinReplicas,
			MaxReplicas:                    config.MaxReplicas,
			TargetCPUUtilizationPercentage: config.TargetCPUUtilizationPercentage,
		},
	}
	_, err := clientset.Autoscaling().HorizontalPodAutoscalers(InstallationNamespace).Create(hpa)
	if err != nil {
		fmt.Printf("Warning: %v\n", err)
		err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
			_, updateErr := clientset.Autoscaling().HorizontalPodAutoscalers(InstallationNamespace).Update(hpa)
			return updateErr
		})
		if err != nil {
			panic(fmt.Errorf("Updating HPA failed: %v", err))
		}
	}
	return err
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
	servicesClient := clientset.CoreV1().Services(ns)
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
	servicesClient := clientset.CoreV1().Services(ns)
	count := 1
	ip := ""
	operation := func() error {
		fmt.Printf("Getting External IP trial %v\n", count)
		count += 1
		currentService, getErr := servicesClient.Get(name, metav1.GetOptions{})
		if getErr != nil {
			// TODO: return error instead of panic
			panic(fmt.Errorf("Failed to get latest version of Service: %v", getErr))
		}
		ingress := currentService.Status.LoadBalancer.Ingress
		if ingress != nil && len(ingress) > 0 {
			ip = ingress[0].IP
			if ip != "" {
				return nil
			}
			ip = ingress[0].Hostname
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
	secretsClient := clientset.CoreV1().Secrets(ns)
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
	secretsClient := clientset.CoreV1().Secrets(ns)
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
