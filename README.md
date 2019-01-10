# vamp lamia command line client

## development
if you have go installed,
git clone it to $GOPATH/src/github.com/magneticio/vamp2cli
so that docker builder works.

## build

If you get errors about missing libraries while building, run:
```
go get
```

for docker build:
```
./build.sh
```
for local build:
```
./build.sh local
```

binaries will be under bin directory

## installation
If you have binaries built locally:
For mac run:
```
./bin/vamp2cli-darwin-amd64 --help
```

If you have downloaded the binary directly, Just copy the binary for you platform to the user binaries folder for general usage, for MacOS:

```
cp vamp2cli-darwin-amd64 /usr/local/bin/vamp2cli
chmod +x /usr/local/bin/vamp2cli
```

If you don't have anything yet and automatically download an install follow commands for your platform:

This installation may not work since this is a private repository.
Manual installation is recommended.

Easy install for MacOS:
```
base=https://github.com/magneticio/vamp2cli/releases/download/0.0.6 &&
  curl -L $base/vamp2cli-$(uname -s)-$(uname -m) >/usr/local/bin/vamp2cli &&
  chmod +x /usr/local/bin/vamp2cli
```

TODO: add installation for other platforms

## Run

Check if it is running with:
```
vamp2cli version
```

This documentation is for version v0.0.5

First you need to login to the vamp application:
You will need
  - the url of your cluster eg.: http://1.2.3.4:8888
  - username eg.: username
  - password eg.: password

Please ask your provider if you don't have these information

If you have installed your vamp into your cluster, these information will be given to you during the installation.
If you are using the SaaS version, this information will be provided by the Vamp.io

```
vamp2cli login --url http://1.2.3.4:8888 --user username --password password
```

For managing all the resources, there is a umbrella concept called project, project name should be unique inside a vamp application. So start by creating a Project.

Please list projects first and choose a new one:
```
vamp2cli list project
```

Starting a new project, users can bundle all the resources under a project.

Please set a project name that is not listed above

```
PROJECT_NAME=myproject
```

This command will let you create an empty project
```
vamp2cli create project $PROJECT_NAME -i json -s "{}"
```

Client allows you to pass specifications as json, yaml and from file or from string. This command reads the input as json and passes the empty json object as configuration. Project doesn't have any mandatory fields, so an empty json is still valid.

You can delete the project just created with:

```
vamp2cli delete project $PROJECT_NAME
```

Please download resources folder from the repo to continue rest of the examples.
Assuming resources folder exists in your current workspace;

Let's create a project from a yaml file currently exists in the resources folder.

```
vamp2cli create project $PROJECT_NAME -f ./resources/project.yaml
```

If you don't have files locally you can also load them from a remote location:

Resources for this readme is localed at https://raw.githubusercontent.com/magneticio/demo-resources/master/vamplamiacliv1


```
vamp2cli create project $PROJECT_NAME -f https://raw.githubusercontent.com/magneticio/demo-resources/master/vamplamiacliv1/project.yaml
```

Rest of the README will use remote resources.

You can check the details of the created project with get method:

```
vamp2cli get project $PROJECT_NAME
```

By default it will print yaml representation of project
For JSON user can specify output, eg.:

```
vamp2cli get project $PROJECT_NAME -o json
```

Let's update the project configuration with an updated file.

```
vamp2cli update project $PROJECT_NAME -f https://raw.githubusercontent.com/magneticio/demo-resources/master/vamplamiacliv1/project_update.yaml
```

Run get project again to see the changes:

```
vamp2cli get project $PROJECT_NAME
```
You will see the second key in the metadata changed to
"key2: value2_new"

This concept applies to every object in vamp system as long as you have access rights as a user.

Vamp has the concept of Project, Cluster, Virtual Cluster in the order of hierarchy.

If you have used kubernetes, Cluster represents configurations related to an actual cluster and Virtual Cluster is configurations bound to a namespace.

You are expected to work in the same project so you can set it as default by running set command:

```
vamp2cli set -p $PROJECT_NAME
```

Now, you can bootstrap your cluster with vamp.
Please make sure you have installed kubectl and authenticated to connect to the cluster you want to be managed by vamp. Command line client will set up a service account user in your cluster and set up credentials to connect to your cluster in vamp.

For this example, it is recommended that you have a cluster of at least 5 nodes, 2 CPU and 8 GB Ram per node. Otherwise you can have pending pods and initialisation will not finish.

```
vamp2cli bootstrap cluster mycluster
```

A simple cluster configuration requires;
- url
- cacertdata
- serviceaccount_token

You can check if your cluster is properly configured by running a get to the cluster.

```
vamp2cli get cluster mycluster
```

with kubectl you can check the namespaces of vamp-system and istio-system and logging is created.

Vamp will run a job in vamp-system namespace to make sure that everything is properly installed and continue running this job until it is finished. Make sure that you have enough resources and there are no pending pods or pods in Crash Loop.

If you have watch command installed, I recommended running following watch command to see instalment in action:
```
watch kubectl get pods --all-namespaces
```

Run kubectl command to see if there are pending pods, which is a common issue when there is not enough resources:

```
kubectl get pods --all-namespaces | grep Pending
```
If there are pending pods after some a few minutes, it is recommended to diagnose the issue, and if it is a resource issue, scale up the cluster.

Again while working on the same cluster it is recommended to set it as default by:

```
vamp2cli set -c mycluster
```

Now it is time to deploy an example application with one version:
```
kubectl apply -f https://raw.githubusercontent.com/magneticio/demo-resources/master/vamplamiacliv1/demo-application-version1.yaml
```

This will create a namespace called vamp-demo and deploy two deployments. There are two ways of importing a namespace to vamp
- label the namespace with "vamp-managed: enabled"
- Create a virtual cluster through vamp

NOTE: Vamp lamia does not create the namespace, it should exist before the creation. If the namespace does not exist, creation will not fail immediately, it will search for namespace and it will delete itself if the namespace does not exist. It is a combination of behaviours that required to scan for namespaces and not managing them.

```
vamp2cli create virtual_cluster vamp-demo -f https://raw.githubusercontent.com/magneticio/demo-resources/master/vamplamiacliv1/virtualcluster.yaml
```

This will re-label the namespace with required settings if the namespace exits. It will not create the namespace, as it is expected to be created by a deployment pipeline.

```
kubectl get ns vamp-demo -o yaml
```

set virtual cluster for ease of use:
```
vamp2cli set -v vamp-demo
```

To expose the application to outside, you will need a gateway:

```
vamp2cli create gateway shop-gateway -f https://raw.githubusercontent.com/magneticio/demo-resources/master/vamplamiacliv1/gateway.yaml
```

Create a destination

```
vamp2cli create destination shop-destination -f https://raw.githubusercontent.com/magneticio/demo-resources/master/vamplamiacliv1/destination.yaml
```

Final step is creating the VampService.
An http load balanced VampService requires the hostnames.
In this example hostname is the IP address, IP address of the gateway is easy to get with get command:

```
vamp2cli get gateway shop-gateway
```
IP is under status > ip,

It is easier to get with a grep command:
```
vamp2cli get gateway shop-gateway | grep ip
```
Or you can use jsonpath parameter to filter the ip, this is an experimental feature and only supports string parameters:

```
vamp2cli get gateway shop-gateway -o=json --jsonpath '$.status.ip'
```

You can also wait until it is available:
```
vamp2cli get gateway shop-gateway -o=json --jsonpath '$.status.ip' --wait
```

To set the IP address to a bash variable run:
DON'T SKIP THIS STEP
```
GATEWAY_IP=$(vamp2cli get gateway shop-gateway -o=json --jsonpath '$.status.ip' --wait)
echo $GATEWAY_IP
```
This command waits until IP is ready and set it to a bash parameter.

You need so see an IP address which is the IP address of the public facing load balancer.

I will reference this ip as GATEWAY_IP in the rest of the documentation.

Please note that this may take some time depending on the provider.

You need to copy paste the IP address of the gateway in the vamp service configuration under the hosts as an array.
Since this is harder to do in remote resources, there is an easier way to add hosts on the fly with a host parameter.

Create a Vamp Service with 100% traffic on version1
```
vamp2cli create vamp_service shop-vamp-service -f https://raw.githubusercontent.com/magneticio/demo-resources/master/vamplamiacliv1/vampservice_template.yaml --host $GATEWAY_IP
```

Check link generated with:
```
echo http://$GATEWAY_IP
```

Now your application is released to public, copy paste the ip of your gateway to your favourite browser (http://$GATEWAY_IP). Now you should be able to see the blue e-commerce page. This is the first version.

Let's deploy the second version of the e-commerce website with kubectl
```
kubectl apply -f https://raw.githubusercontent.com/magneticio/demo-resources/master/vamplamiacliv1/demo-application-version2.yaml
```

Now with this canary release, the second version will be released in time based manner,
If there is an error in the responses, it will automatically roll back and retry until successful.

You can create a canary release with configuration file or you can use the release command.

Create with configuration:
```
vamp2cli create canary_release shop-release -f https://raw.githubusercontent.com/magneticio/demo-resources/master/vamplamiacliv1/canaryrelease.yaml
```

Release command:
```
vamp2cli release shop-vamp-service --destination shop-destination --subset subset2 -l version=version2
```

Check your browser and refresh frequently to see the second version is available.

You can also check the percentage changes with:
```
vamp2cli get vamp_service shop-vamp-service
```

It will take some time to release totally and you can not see the first version anymore.

But now you decided, a url based access to these version are more useful for you, then you can set up conditional routes.

Before going forward delete current canary release
```
vamp2cli list canary_release
vamp2cli delete canary_release shop-vamp-service-shop-destination-subset2
```


Update the hosts field the IP address of gateway in ./resources/conditionalvampservice.yaml then update the vamp service:

```
vamp2cli update vamp_service shop-vamp-service -f https://raw.githubusercontent.com/magneticio/demo-resources/master/vamplamiacliv1/conditionalvampservice_template.yaml --host $GATEWAY_IP
```

Now you can see that first version is available under url:
http://$GATEWAY_IP/v1
and second version is under
http://$GATEWAY_IP/v2

Check links generated with:
```
echo http://$GATEWAY_IP/v1
echo http://$GATEWAY_IP/v2
```
