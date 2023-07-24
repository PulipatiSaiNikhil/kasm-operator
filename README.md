# The Idea of Replicating Kasm Workspace in Openshift

## For this, we create a Go-based Operator using Operator-sdk CLI

### Prerequisites:

1. Operator-sdk (CLI).

2. Openshift oc utility.

3. Kubectl binary.

4. Setup GVM and use go v1.20.

5. (Optional) VS code and SSH extension to connect your VM remotely.

6. Logged into an OCP cluster with `oc` with an account with cluster-admin permissions.

   <br>

### Installing Prerequisites:

#### 1.Setup GVM and use go v1.20

Installing GVM is straightforward, but before installing GVM make sure to install it's dependencies such as `curl`, `git`, `mercurial`, `make`, `binutils`, `bison`, `gcc` and `build-essential`

To install all the all the dependencies mentioned above run the following command:

`sudo apt-get install curl git mercurial make binutils bison gcc build-essential `

The [GVM repository](https://github.com/moovweb/gvm#installing) installation documentation instructs you to download the installer  script and pipe it to Bash:

```bash
bash < <(curl -s -S -L https://raw.githubusercontent.com/moovweb/gvm/master/binscripts/gvm-installer)
```

 After installing GVM, we can manage go versions and install them using GVM.

 The command `gvm listall` gives all the versions of go and the specific go version can be installed by the command `gvm install <version>`  where `<version>` is one of those returned by `gvm listall` command.

After installing the desired version of `go` use the command `gvm use <version>` to use the go with the version mentioned.

Use the command `rm -rf ~/.gvm` to remove GVM.

#### 2.Operator-sdk CLI Installation(on Linux)

Prerequisites for installing Operator-sdk CLI is to have a `Go` v1.19+ and `docker` v1.17+ or `podman` 1.9.3+

Navigate to the [OpenShift mirror site](https://mirror.openshift.com/pub/openshift-v4/x86_64/clients/operator-sdk/) and from the latest 4.12 directory, download the latest version of tarball for Linux.

Unpack the archive and make the file executable using the following commands:

```bash
tar xvf operator-sdk-v1.25.4-ocp-linux-x86_64.tar.gz (this is one of the version of the operator-sdk-v1.25.4)

chmod +x operator-sdk
```

Move the extracted `operator-sdk` binary to a directory that is on your `PATH`(To check `PATH` use `echo $PATH` command)

```bash
sudo mv ./operator-sdk /usr/local/bin/operator-sdk
```

After installing the `operator-sdk CLI` verify that it is available:

`operator-sdk version`  command gives the output as version of the `operator-sdk` CLI installed

#### 3.Openshift oc utility

Update the package manager's cache by running the following command:

```bash
sudo apt update
```

Install the `curl` package if it's not already installed. The `curl` command is used to download files from the internet.

```bash
sudo apt install curl
```

Download the oc CLI binary using curl.

```bash
curl -LO https://mirror.openshift.com/pub/openshift-v4/clients/oc/latest/linux/oc.tar.gz
```

Extract the downloaded archive using the tar command.

```bash
tar -xvf oc.tar.gz
```

Move the oc binary to a directory in your system's PATH,such as `/usr/local/bin`.

```bash
sudo mv oc /usr/local/bin/
```

Verify that the oc CLI is installed properly by running the following command:

```bash
oc version
```

#### 4.kubectl binary

Download the latest version of `kubectl` using curl.

````bash
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
````

Make the downloaded binary executable by running the following command:

```bash
chmod +x kubectl
```

Move the `kubectl` binary to a directory included in your system's PATH, such as `/usr/local/bin`.

```bash
sudo mv kubectl /usr/local/bin/
```

Verify the `kubectl` is installed properly by running the following command:

```bash
kubectl version --client
```

#### 5.(Optional)VS code and SSH extension to to connect your VM remotely

If you are using **WSL(Windows Subsystem for Linux)**, we can use SSH(Secure Shell) to connect to a remote server/system securely using VS Code.

##### Establishing Remote Connection

With the Remote-SSH extension installed, you will see a new Status bar item at the bottom left of VS Code, as shown in the below image

![../operator/vs code/Screenshot1.png](../operator/vs code/Screenshot1.png)

The __Remote Status bar__ item can quickly show you in which context VS Code is running (local or remote) and clicking on the item will bring up the options as shown below

![../operator/vs code/Screenshot2.png](../operator/vs code/Screenshot2.png)

As we are WSL in this documentation, we can either choose options **Connect to WSL** or **Connect to WSL using Distro...**

If Connect to WSL is chosen, the default WSL would be connected else if Connect to WSL using Distro... is chosen, a dropdown of all the WSL present in the system would be displayed and can connect to any of the subsystem as shown below.

![../operator/vs code/Screenshot3.png](../operator/vs code/Screenshot3.png)

After selecting the distro, VS Code securely connects to the remote server/system through SSH.

![../operator/vs code/Screenshot4.png](../operator/vs code/Screenshot4.png)

The above image shows that kali-Linux remote system is connected to VS Code through SSH and can develop from here and access the terminal from terminal=>new terminal option in VS Code.

##### Closing Remote Connection

If you wish to close the connection with the remote server/system, click on the Remote status bar which shows up options as below

![../operator/vs code/Screenshot5.png](../operator/vs code/Screenshot5.png)

Choose **Close Remote Connection** to end the remote server/system connection.

#### 6. Create an user using oc and provide cluster-admin permissions

By default, only a `kubeadmin` user exists on your cluster. and we can create an user using htpasswd authentication in OCP.

Using htpasswd authentication in Openshift Container Platform allows you to identify users based on an htpasswd file. An htpasswd file is a flat file that contains the user name and hashed password for each user. You can use `htpasswd` utility to create this file.

Create your flat file with a user name and hashed password:

``` bash
htpasswd -c -B -b /tmp/htpasswd <username> <password>
```

The command generates a hashed version of the password.

here,

`-c` indicates to create a new flat file.

`-B` indicates to use `bcrypt` encryption for passwords.

`-b` Use batch mode; *i.e.*, get the password from the command line rather than prompting for it. This option should be used with extreme care, since **the password is clearly visible** on the command line.

we can use `-i` to read the password from stdin without verification (for script usage).

Continue to add or update credentials to the file:

```bash
htpasswd -B -b /tmp/htpasswd <user_name> <password>
```

Login to the cluster with username `kubeadmin` as that default user has the cluster-admin permissions.

##### Creating the htpasswd secret

To use the htpasswd identity provider, you must define a secret that contains the `htpasswd` user file.

Use the following command to create a `secret` object whcih contains the `htpasswd` users file:

``` bash
oc create secret generic htpass-secret --from-file=htpasswd=/tmp/htpasswd -n openshift-config
```

check for the manifest of the secret created using the following command:

``` bash
oc get secret htpass-secret -n openshift-config -o yaml
```

after creating the secret, edit the existing `OAuth` CR to add the secret to the spec section of the CR.

using following command we can add the above `htpass-secret` `Secret` object to the CR:

```bash
oc edit oauth secret
```

and in the spec section add the following and save it.

```bash
spec:
  identityProviders:
  - name: my_htpasswd_provider
    mappingMethod: claim
    type: HTPasswd
    htpasswd:
      fileData:
        name: htpass-secret(secret object name)
```

after saving it, the changes automatically occur and the users get added to the cluster and now you can use `oc login` command to login to cluster using the usernames provided in htpasswd file.

To give the user cluster-admin privileges use the following command:

```bash
oc adm policy add-role-to-user cluster-admin <username>
```



If you wish to remove a particular user then run the following command:

```bash
htpasswd -D /tmp/htpasswd <username>
```

and update the existing `htpass-secret` `Secret` object with the updated users in the htpasswd file using the following command:

```bash
oc create secret generic htpass-secret --from-file=htpasswd=/tmp/htpasswd --dry-run=client -o yaml -n openshift-config | oc replace -f -
```

If you remove one or more users, you must additionally remove existing resources for each user. i.e. Delete the `user` object and Delete the `Identity` object for the user.

```bash
oc delete user <username>
oc delete identity my_htpasswd_provider:<username>
```

<br>

## Creating a Project

Create a directory for the project:

```bash
mkdir $HOME/kasmmod
```

Change to the directory:

```bash
cd $HOME/kasmmod
```

Activate support for Go modules:

```bash
export GO111MODULE=on
```

Run the `operator-sdk init` command to initialize the project:

```bash
operator-sdk init --domain=<your-domain> --repo=<git-repo>
```

For this operator the command used is:

```bash
operator-sdk init --domain=kasmmod.office.ocpztp.com --repo=github.com/balakuberox/kasmmod.git
```

make the git repo public, which is mentioned here, so it can be used again in the future.

(The `operator-sdk init` command generates a `go.mod` file for go modules. )

<br>

## PROJECT file

Among the files generated by the `operator-sdk init` command is a Kubebuilder  `PROJECT` file.

This file represents the project’s configuration and is used to track useful information for the CLI and plugins.

<br>

## About multi-group APIs

Before you create an API and controller, consider whether your Operator requires multiple API groups.

To change the layout of your project to support multi-group APIs, you can run the following command

```bash
operator-sdk edit --multigroup=true
```

This command updates the `PROJECT file`, which adds the line `multigroup: true`.

<br>

## Creating an API and controller

Use the Operator SDK CLI to create a custom resource definition (CRD) API and controller.

Run the following command to create an API group `cache` , version `v1`, and kind of any choice(that you wish to have).

```bash
operator-sdk create api --group=ocpztp --version=v1 --kind=kasmmod
```

kind - name of the custom resource that you wanted to have in the cluster.

after hitting enter, enter y for creating `resource` and `controller` when prompted and that will create a scaffold for you to edit.

<br>

## Defining the API

Modify the Go type definition at `api/v1/kasmmod_types.go` to have the following `spec` and `status`:

<img src="../operator/Spec_status.jpeg" alt="spec_status.png" style="zoom: 50%;" />

Update the generated code with for the resource type:

```bash
make generate
```

The above Makefile target invokes the controller-gen utility to update `api/v1/zz_generated.deepcopy.go ` file. This ensures your API Go type definitions implement the `runtime.Object` interface that all Kind types must implement.

<br>

## Generating CRD manifests

After the API is defined with `spec` and `status` fields and custom resource definition (CRD) validation markers, you can generate CRD manifests.

The following command generate and update CRD manifests:

```bash
make manifests
```

This Makefile target invokes the `controller-gen` utility to generate the CRD manifests in the `config/crd/bases/<group>.<domain>_kasmmods.yaml` file.

<br>

## Implementing the controller

You can implement that controller logic after creating a new API and controller.

For this Operator, update the generated controller file `controllers/kasmmod_controller.go` that created a service account, Role binding, deployment, service and routes.

The below code snippet explains about how the deployment is created.

![./deployment.png](../operator/deployment.png)

```go
found := &appsv1.Deployment{}
```

- This line declares a variable `found` of type `*appsv1.Deployment` (a pointer to `appsv1.Deployment` struct) and assigns an empty instance of `appsv1.Deployment` to it.

```
err = r.Get(ctx, types.NamespacedName{Name: Kasmmod.Name, Namespace: Kasmmod.Namespace}, found)
```

- This line attempts to retrieve a deployment resource using the Kubernetes client's `Get` method. It takes three arguments:
  - `ctx` is a context object for the request.
  - `types.NamespacedName{Name: Kasmmod.Name, Namespace: Kasmmod.Namespace}` specifies the name and namespace of the deployment to retrieve.
  - `found` is a pointer to the `appsv1.Deployment` object where the retrieved deployment will be stored.
- The potential error is assigned to the `err` variable.

```go
if err != nil && errors.IsNotFound(err) { ... } else if err != nil { ... }
```

- This block of code checks the value of `err` to determine the outcome of the `Get` operation.
- If `err` is not `nil` and `errors.IsNotFound(err)` returns `true`, meaning the deployment was not found, and the code creates a new deployment.
- If `err` is not `nil` but the `errors.IsNotFound(err)` check fails, it means an error occurred during the `Get` operation, and the code logs the error and returns the error as the result.

Going inside `if` part of the code snippet

```go
dep := r.deploymentForKasmmod(Kasmmod)
```

- This line calls a method `deploymentForKasmmod`. It passes the `Kasmmod` object as an argument.
- This method aims to generate a new `Deployment` object based on the `Kasmmod` object. It likely contains the logic to set the desired deployment specifications.
- The below function defines the creation of a new `Deployment` using `deploymentForKasmmod` function.

<img src="../operator/deployment_func.png" alt="deployment_func.png" style="zoom:50%;" />

```
log.Info("Creating a new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
```

- This line logs an informational message indicating a new deployment is being created. It includes the namespace and name of the deployment (`dep`) as additional information.

```go
err = r.Create(ctx, dep)
```

- This line attempts to create a new deployment using the Kubernetes client's `Create` method. It takes two arguments:
  - `ctx` is a context object for the request.
  - `dep` is the `Deployment` object to be created.
- The potential error is assigned to the `err` variable.

```go
else if err != nil { ... }
```

- This block of code checks if the deployment creation resulted in an error.
- If `err` is not `nil`, it means an error occurred during the creation, and the code logs the error and returns the error as the result.

As mentioned above, replicate the same for `service account`, `role binding`, `service` and `route` and refer the code if there are issues or refer this page [Go Packages](https://pkg.go.dev/) for any package related issues (mainly the issue may be of the incorrect package usage in the code or some syntax error which can be corrected).

While writing the functions for creation of `service account`, `role binding`, `service` and `route` make sure that you import the packages required for the creation of the above.



As mentioned the status of the deployment contains the pod names, the following code snippet updates the status of a `Kasmmod` resource with the names of associated pods.

![status.png](../operator/status.png)

```go
podList := &corev1.PodList{}
```

- This line declares a variable `podList` of type `corev1.PodList` and assigns an empty instance of `corev1.PodList` to it. This object will be used to store the list of pods related to the `Kasmmod` resource.

```go
listOpts := []client.ListOption{ ... }
```

- This line declares a variable `listOpts` as a slice of `client.ListOption`. It defines options for listing pods related to the `Kasmmod` resource.

- The options include:

  - `client.InNamespace(Kasmmod.Namespace)`: This option filters the pods to only those in the same namespace as the `Kasmmod` resource.
  - `client.MatchingLabels(labelsForKasmmod(Kasmmod.Name))`: This option filters the pods based on the labels that match the labels returned by the `labelsForKasmmod` function, which likely generates labels specific to the `Kasmmod` resource.

     The following code has the function `labelsForKasmmod` which takes the argument as a `string` and returns a map.

     <img src="../operator/labels.png" alt="status.png" style="zoom: 50%;" />

```go
if err = r.List(ctx, podList, listOpts...); err != nil { ... }
```

- This block of code uses the Kubernetes client's `List` method to retrieve a list of pods based on the provided options.
- The `r.List` method takes three arguments:
  - `ctx` is a context object for the request.
  - `podList` is the object where the retrieved list of pods will be stored.
  - `listOpts...` spreads the `listOpts` slice into individual arguments.
- If an error occurs during the listing operation, the error is logged, and the function returns the error as the result.

```go
podNames := getPodNames(podList.Items)
```

- This line calls a function `getPodNames` and passes the `Items` property of the `podList` object as an argument.
- The purpose of the `getPodNames` function is to extract the names of the pods from the `Items` list and return them as a slice of strings.
- The extracted pod names will be stored in the `podNames` variable.

```go
if !reflect.DeepEqual(podNames, Kasmmod.Status.Nodes) { ... }
```

- This condition checks if the `podNames` slice obtained in the previous step is not equal to the `Nodes` field in the `Status` of the `Kasmmod` resource.
- `reflect.DeepEqual` is used to compare the equality of the two slices, considering their elements and order.
- If the condition is true, it means the pod names have changed, and the status needs to be updated.

```go
Kasmmod.Status.Nodes = podNames
```

- This line updates the `Nodes` field in the `Status` of the `Kasmmod` resource with the `podNames` slice obtained earlier.

```
err := r.Status().Update(ctx, Kasmmod)
```

- This line uses the Kubernetes client's `Status().Update` method to update the status of the `Kasmmod` resource.
- The `Update` method takes two arguments:
  - `ctx` is a context object for the request.
  - `Kasmmod` is the modified `Kasmmod` resource with updated status.
- If an error occurs during the update operation, the error is logged, and the function returns the error as the result.

<br>

## Creating SCC manifest and Rolebinding with service account

In this section, we discuss about creating the SCC manifest with minimal options which give the privileges to run the Kasm application in Openshift and then binding the service account to that SCC.

After the creation of SCC manifest, create the SCC with the command `oc create -f <scc-manifest>`

Here is an example of the SCC manifest that has the privileges to run Kasm application

<img src="../operator/scc.png" alt="./scc.png" style="zoom:50%;" />

`allowPrivilegedContainer: false`

 This line indicates whether privileged containers are allowed. Privileged containers have more privileges and can perform actions that are typically restricted for security reasons. In this case, privileged containers are not allowed.

`readOnlyRootFilesystem: false`

This line specifies whether the container's root filesystem is read-only or not. If set to true, the container's root filesystem is read-only. In this case, it is set to false.

`priority: 50`

This line sets the priority of the SCC. The priority determines the order in which SCCs are evaluated when assigning them to pods. Higher priority values are evaluated first.

`runAsUser`: This section specifies the constraints related to the user under which the container runs.

- `type: RunAsAny`: This line indicates that the container can run as any user. It does not enforce any specific user ID restrictions.

`seLinuxContext`: This section specifies the constraints related to SELinux(Security Enhanced Linux) context.

- `type: RunAsAny`: This line indicates that the container can run with any SELinux context. It does not enforce any specific SELinux context restrictions.

`volumes`: This section lists the allowed volume types for the pod.



Create the SCC from SCC manifest using the following command:

```bash
oc create -f <scc-manifest>
```

Check whether the SCC is reflected in the list of SCC by using the following command:

```bash
oc get scc
```

Now we should bind the service account with the SCC, which is done by the `reconcile` function in the `controller` file and the `rolebinding ` object for `Kasmmod` resource looks similar to this:

<img src="../operator/rb_func.png" alt="./rb_func.png" style="zoom: 50%;" />

`Subjects: []v1.Subject{...}`:

This section specifies the subjects associated with the RoleBinding. In this case, there is a single subject of kind "ServiceAccount" specified by the `m.Spec.Serviceaccount` field of the `Kasmmod` resource. The subject's Name and Namespace are also set based on the `Kasmmod` resource.

In the `RoleRef: v1.RoleRef{...}` we mention the `APIGroup`, `Kind` and `Name` of the SCC to which is rolebinded to.

**`ctrl.SetControllerReference(m, rb, r.Scheme)`** sets the `Kasmmod` resource (`m`) as the owner of the Rolebinding object (`rb`). This ensures that when the `Kasmmod` resource is deleted, the associated RoleBinding object will also be deleted.

This would create an individual RoleBinding for each service account rolebinded to SCC with RoleBinding name same as that of the service account.

If the `service account` is not provided in the `Kasmmod` CR, the default `serviceaccount` is used and no RoleBinding would be created with the default `serivceaccount` .



<br>

## Running the Operator

### Running locally outside the cluster

This is useful for development purposes to speed up deployment and testing.

Run the following command to install the custom resource definition(CRDs) in the cluster configures in your `~/.kube/config` file and run the Operator locally:

```bash
make install run
```

make sure you have the `kubectl` utility before you run the command.

As mentioned above for controller, if any changes are done to `api/v1/kasmmod_types.go`, we use `make manifests` to update the CRD manifests.

Now the Kasmmod Operator, which provides Kasmmod  CR, installed on the cluster.

<br>

## Creating the Custom Resource file

We try to create the CR by referring to `api/v1/kasmmod_types.go`

In the `api/v1/kasmmod_types.go` file let's have a look at the `spec` part:

<img src="../operator/Spec.png" alt="Spec.png" style="zoom:50%;" />

The `omitempty` tag ensures that if the field is not set in the JSON representation, it will be omitted.

So, only the `Size` of the `Kasmmod` resource is compulsory to be provided in the CR manifest and the others have a default value included in the `controller` file.

The `Sessionid` refers to the name of the router which is used to create the host using `routes` in `oc`.

An Example of the Custom Resource of kind `Kasmmod` is shown below.

<img src="../operator/CR.png" alt="CR.png" style="zoom:67%;" />

The `apiVersion` in the CR is broken down into `<group>.<domain>/<api-version>`

```bash
 operator-sdk init --domain=<domain> --repo=github.com/example-inc/memcached-operator
```

the `<domain>` in the `apiVersion` resembles to the `domain` in the `operator-sdk init` command.

```bash
operator-sdk create api --group=cache --version=v1 --kind=Kasmmod
```

The `<group>` and `<api-version>` in the `apiVersion` resembles to the `group` and `version` in the `operator-sdk create api` command.

From above example CR, `domain=kasmmod.office.ocpztp.com` , `group=ocpztp` and, `version=v1`

To create the CR, run the following command:

```bash
oc apply -f <CR-manifest-file
```

or

```bash
oc create -f <CR-manifest-file>
```

The above command will check the CR with the provided name in the manifest is created or not, if not created it creates a CR with the name provided in the manifest else if the CR with that name already exists, then it updates the current deployment according to the manifest file included in `oc` command.

The command `oc create` just creates the CR and you can't update the CR using the `oc create` command.

**Note:** The `apply` only updates the deployment if there is a new entry in the manifest, if you edit the existing values in the manifest and use the `oc` command that will not update the deployment and it is advised  to delete the CR and deploy it again.

To delete a CR, run the following command:

```bash
oc delete kasmmod <CR-name>
```

or

``` bash
oc delete -f <CR-manifests-file>
```

<br>

## Code References

GitHub repo link: [https://github.com/balakuberox/kasmmod](https://github.com/balakuberox/kasmmod)

<br>

## Creating another CR for modifying the Kasmmod CR

### Purpose of creating another CR

As the customer cannot directly interact with the `Kasmmod` CR and change the fields in the CR, creating a new CR that can change the existing `Kasmmod` CR would be a feasible solution for the problem of interacting directly with the `Kasmmod` CR.

Let's name the new CR as **KasmmodTemplate**.

Now use of the multi-group APIs to create an API with a group name the same as that of `Kasmmod`(i.e.`ocpztp`), version as ` v1alpha1` and kind as `KasmmodTemplate`.

**Note**: It is advised to use different API versions to help distinguish between different versions of API and controller files.

The command to create a new API with type KasmmodTemplate:

```bash
operator-sdk create api --group=ocpztp --version=v1alpha1 --kind=KasmmodTemplate
```

after hitting enter, enter y for creating `resource` and `controller` when prompted and that will create a scaffold for you to edit of version `v1aplha1`.

Reorder the API and Controller files into folders named `v1` and `v1alpha1`

**Note**: After reordering the files, edit the corresponding files, such as `main.go`, `controller file`, which may give errors in the packages, and the reason might be the incorrect path of the files. So make sure that the path specified in the packages matches the exact path of the corresponding file.

### Defining the API for KasmmodTemplate

Modify the Go type definition at `api/v1alpha1/kasmmodtemplate_types.go` to have the following `spec` and `status`:

<img src="../operator/kasmmodtemplate_api.png" alt="../operator/kasmmodtemplate_api.png" style="zoom:50%;" />

Here we use the field `TargetKasmod` as mandatory field to be provided in `KasmmodTemplate` CR which takes the value as name of the existing `kasmmod` CR on the cluster .

Let's assume the `Kasmmod` CR name as `sample1` as an example. The `sample1` CR is modified with the values specified in the `KasmmodTemplate` CR and invokes the `Reconcile` function in `Kasmmod` controller which updates the deployment accordingly.

Update the generated code with for the resource type:

```bash
make generate
```

The above Makefile target invokes the controller-gen utility to update `api/v1alpha1/zz_generated.deepcopy.go ` file. This ensures your API Go type definitions implement the `runtime.Object` interface that all Kind types must implement.

### Generating CRD manifests for KasmmodTemplate

After the API is defined with `spec` and `status` fields and custom resource definition (CRD) validation markers, you can generate CRD manifests.

The following command generate and update CRD manifests:

```bash
make manifests
```

This Makefile target invokes the `controller-gen` utility to generate the CRD manifests in the `config/crd/bases/cache.<domain-name>_kasmmodtemplates.yaml` file.

### Implementing the controller for KasmmodTemplate

Using the `KasmmodTemplate` CR, we don't deploy any deployments, but this is used to modify the deployments of `Kasmmod` by changing the values in `Kasmmod` CR with the corresponding name specified in the `KasmmodTemplate` CR.

Now change the code so that the `Kasmmod` CR values will be changed corresponding to the `TargetKasmmod` name specified in `KasmmodTemplate` CR.

The controller code for the `Kasmmod` and `KasmmodTemplate` are at paths `/controllers/v1/kasmmod_controller.go` and `/controllers/v1alpha1/kasmmodtemplate_controller.go`, respectively.



## Some IMP points to be noted while using KasmmodTemplate

Here are the points to be noted while modifying the controller code of both **`Kasmmod`** and **`KasmmodTemplate`**.

1. When the `port` number of the `Kasmmod` CR is updated through `KasmmodTemplate` CR, the service `port` and `targetport` should also be updated with the new `port` number mentioned in `KasmmodTemplate` CR.

2. Let's have a scenario where we didn't mention the service account field in Kasmmod CR, and that deployment takes the `default` service account. In the `KasmmodTemplate` CR, we mention the name of the `Kasmmod` CR which was deployed earlier, and has a service account field.

   Now check for the service account mentioned in `KasmmodTemplate` CR and whether it exists in that `namespace`. If the service account exists, use that service account and the rolebinding associated with it. If the rolebinding doesn't exist, create a rolebinding and use it, and if the service account doesn't exist, create a new service account with that name and create a rolebinding with the SCC and use that service account for the deployment.
