# vamp lamia command line client

## development
if you have go installed,
git clone it to $GOPATH/src/github.com/magneticio/vamp2cli
so that docker builder works.

## build

for docker build:
```
./build.sh
```
for local build:
```
./build.sh local
```

binaries will be under bin directory

## Run
For mac run:
```
./bin/vamp2cli-darwin-amd64 --help
```

If you have downloaded the binary directly, Just copy the binary for you platform to the user binaries folder for general usage, for MacOS:

```
cp vamp2cli-darwin-amd64 /usr/local/bin/vamp2cli
chmod +x /usr/local/bin/vamp2cli
```

Check if it is running with:
```
vamp2cli version
```

First you need to login to the vamp application:
You will need
  - the url of your cluster eg.: http://1.2.3.4:8888
  - username eg.: username
  - password eg.: password

If you have installed your vamp into your cluster, these information will be given to you during the installation.
If you are using the SaaS version, this information will be provided by the Vamp.io

```
vamp2cli --url http://1.2.3.4:8888 --user username --password password
```

For managing all the resources, there is a umbrella concept called project, project name should be unique inside a vamp application. So start by creating a Project.

This command will let you create an empty project
```
vamp2cli create project myproject -i json -s "{}"
```

Client allows you to pass specifications as json, yaml and from file or from string. This command reads the input as json and passes the empty json object as configuration. Project doesn't have any mandatory fields, so an empty json is still valid.

You can delete the project just created with:

```
vamp2cli delete project myproject
```

Please download resources folder from the repo to continue rest of the examples.
Assuming resources folder exists in your current workspace;

Let's create a project from a yaml file currently exists in the resources folder.

```
vamp2cli delete project myproject -f ./resources/project.yaml
```

You can check the details of the created project with get method:

```
vamp2cli get project myproject
```

By default it will print yaml representation of project
For JSON user can specify output, eg.:

```
vamp2cli get project myproject -o json
```

Let's update the project configuration with an updated file.

```
vamp2cli update project myproject -f ./resources/project_update.yaml
```

Run get project again to see the changes:

```
vamp2cli get project myproject
```
You will see the second key in the metadata changed to
"key2: value2_new"

This concept applies to every object in vamp system as long as you have access rights as a user.

Vamp has the concept of Project, Cluster, Virtual Cluster in the order of hierarchy.

If you have used kubernetes, Cluster represents configurations related to an actual cluster and Virtual Cluster is configurations bound to a namespace.

You are expected to work in the same project so you can set it as default by running set command:

```
vamp2cli set -p myproject
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

Run kubectl command to see if there are pending pods, which is a common issue when there is not enough resources:

```
kubectl get pods --all-namespaces | grep Pending
```
If there are pending pods after some a few minutes, it is recommended to diagnose the issue, and if it is a resource issue, scale up the cluster.

Again while working on the same cluster it is recommended to set it as default by:

```
vamp2cli set -c mycluster
```

Now it is time to deploy an example application with two versions:
```
kubectl apply -f ./resources/demo-application.yaml
```

This will create a namespace called vamp-demo and deploy two deployments. There are two ways of importing a namespace to vamp
- label the namespace with "vamp-enable" : "true"
- Create a virtual cluster through vamp

```
vamp2cli create virtual_cluster vamp-demo -f ./resources/virtualcluster.yaml
```

This will re-label the namespace with required settings if the namespace exits. It will not create the namespace, as it is expected to be created by a deployment pipeline.


To expose the application to outside, you will need a gateway:

```
vamp2cli create gateway shop-gateway -f ./resources/gateway.yaml
```

Create a destination

```
vamp2cli create destination shop-destination -f ./resources/destination.yaml
```

Final step is creating the VampSerivice.
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
Copy paste the IP address of the gateway in the vamp service configuration under the hosts, the resource file It is 1.2.3.4 by default.

Create a Vamp Service with 50%-50% traffic
```
vamp2cli create vamp_service shop-vamp-service -f ./resources/vampservice.yaml
```
