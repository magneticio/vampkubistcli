package kubeclient

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	// Uncomment the following line to load the gcp plugin (only required to authenticate against GKE clusters).
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

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

func BootstrapVampService() (string, string, string, error) {
	// create the clientset
	clientset, host, err := getLocalKubeClient()
	if err != nil {
		panic(err.Error())
	}
	ns := "vamp-system"
	nsSpec := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}}

	_, err_n := clientset.Core().Namespaces().Create(nsSpec)
	if err_n != nil {
		// panic(err_n.Error())
		fmt.Printf("Warning: %v\n", err_n.Error())
	}

	/*
	   kind: ClusterRoleBinding
	   apiVersion: rbac.authorization.k8s.io/v1
	   metadata:
	     name: vamp-sa-cluster-admin-binding
	   subjects:
	   - kind: User
	     name: system:serviceaccount:vamp-system:default
	     apiGroup: rbac.authorization.k8s.io
	   roleRef:
	     kind: ClusterRole
	     name: cluster-admin
	     apiGroup: rbac.authorization.k8s.io
	*/

	clusterRoleBindingSpec := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{Name: ns + "-sa-cluster-admin-binding"},
		Subjects:   []rbacv1.Subject{rbacv1.Subject{Kind: "ServiceAccount", Name: "default", Namespace: ns}},
		RoleRef:    rbacv1.RoleRef{Kind: "ClusterRole", Name: "cluster-admin", APIGroup: ""},
	}
	_, err_c := clientset.RbacV1().ClusterRoleBindings().Create(clusterRoleBindingSpec)

	if err_c != nil {
		// panic(err_n.Error())
		fmt.Printf("Warning: %v\n", err_c.Error())
	}

	/*
	  apiVersion: v1
	  kind: Secret
	  metadata:
	    name: vamp2imagepull
	    namespace: vamp-system
	  type: kubernetes.io/dockercfg
	  data:
	     .dockercfg: eyJodHRwczovL2luZGV4LmRvY2tlci5pby92MS8iOnsiYXV0aCI6ImRtRnRjREp3ZFd4c09uWmhiWEF5Y0hWc2JFWnNkWGc9In19
	*/

	// this should be a variable
	secretDataString := "{\"https://index.docker.io/v1/\":{\"auth\":\"dmFtcDJwdWxsOnZhbXAycHVsbEZsdXg=\"}}"

	pullSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: ns + "vamp2imagepull", Namespace: ns},
		Data: map[string][]byte{
			".dockercfg": []byte(secretDataString),
		},
		Type: "kubernetes.io/dockercfg",
	}

	_, err_s := clientset.Core().Secrets(ns).Create(pullSecret)
	if err_s != nil {
		// panic(err_n.Error())
		fmt.Printf("Warning: %v\n", err_s.Error())
	}

	// This is end of setting up remote vamp set up
	// now we need to get information to connect to the cluster

	/*
		res, err_d := clientset.SettingsV1alpha1()
		if err_d != nil {
			// panic(err_n.Error())
			fmt.Printf("Warning: %v\n", err_d.Error())
		} else {
			fmt.Printf("Response: %v\n", res)
		}
	*/

	// res := clientset.SettingsV1alpha1().RESTClient().Config.Host
	// fmt.Printf("Response: %v\n", res)
	// fmt.Printf("Host: %v\n", host)

	getOptions := metav1.GetOptions{}
	sa, err_sa := clientset.Core().ServiceAccounts(ns).Get("default", getOptions)
	if err_sa != nil {
		// panic(err_n.Error())
		fmt.Printf("Warning: %v\n", err_s.Error())
	}
	// fmt.Printf("Sa Secret name: %v\n", sa.Secrets[0].Name)

	saSecret, err_sa_secret := clientset.Core().Secrets(ns).Get(sa.Secrets[0].Name, getOptions)
	if err_sa_secret != nil {
		// panic(err_n.Error())
		fmt.Printf("Warning: %v\n", err_sa_secret.Error())
	}
	// fmt.Printf("Sa Secret: %v\n", saSecret)
	crt := string(saSecret.Data["ca.crt"])
	// fmt.Printf("Sa Secret crt: %v\n", crt)
	token := string(saSecret.Data["token"])
	// fmt.Printf("Sa Secret token: %v\n", token)

	return host, crt, token, nil
}

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
