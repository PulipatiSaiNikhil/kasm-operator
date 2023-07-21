/*
Copyright 2020.

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

package controllers

import (
	"bytes"
	"context"

	//"fmt"
	"os"
	"os/exec"
	"reflect"
	"strings"

	ocpztpv1 "github.com/balakuberox/kasmmod.git/api/v1"
	"gopkg.in/yaml.v2"

	//ocpztpv1alpha1 "github.com/balakuberox/kasmmod.git/api/v1alpha1"
	"github.com/go-logr/logr"
	routev1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

// KasmmodReconciler reconciles a Kasmmod object
type KasmmodReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=ocpztp.example.com,resources=Kasmmods,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=ocpztp.example.com,resources=Kasmmods/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=ocpztp.example.com,resources=Kasmmods/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Kasmmod object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.0/pkg/reconcile
func (r *KasmmodReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrllog.FromContext(ctx)
	// Fetch the Kasmmod instance
	Kasmmod := &ocpztpv1.Kasmmod{}
	err := r.Get(ctx, req.NamespacedName, Kasmmod)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("Kasmmod resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get Kasmmod")
		return ctrl.Result{}, err
	}

	// Check if the service account already exists, if not create a new one
	// fmt.Printf(" starting of the sa data of Kasmmod.Spec.Serviceaccount , %s", Kasmmod.Spec.Serviceaccount)
	if Kasmmod.Spec.Serviceaccount != "" {
		foundsa := &corev1.ServiceAccount{}
		err = r.Get(ctx, types.NamespacedName{Name: Kasmmod.Spec.Serviceaccount, Namespace: Kasmmod.Namespace}, foundsa)
		//has not error which means there is a deployment already exist
		if err != nil && errors.IsNotFound(err) {
			sa := r.ServiceaccountForKasmmod(Kasmmod, Kasmmod.Spec.Serviceaccount)
			log.Info("Creating a new service account", "service.Namespace", sa.Namespace, "service.Name", sa.Name)
			err = r.Create(ctx, sa)
			if err != nil {
				log.Error(err, "Failed to create new service account", "Service.Namespace", sa.Namespace, "Service.Name", sa.Name)
				return ctrl.Result{}, err
			}
			log.Info("sucessfuly created a new service account", "service.Namespace", sa.Namespace, "service.Name", sa.Name)
			//check if the rolebinding is avaiable
			/*This line initializes a new RoleBinding object using the v1 API version.
			It creates an empty RoleBinding instance that will be used to store the retrieved data.*/

			foundrb := &v1.RoleBinding{}
			err = r.Get(ctx, types.NamespacedName{Name: Kasmmod.Spec.Serviceaccount, Namespace: Kasmmod.Namespace}, foundrb)
			if err != nil && errors.IsNotFound(err) {
				newrb := r.RolebindingcreationForKasmmod(Kasmmod, Kasmmod.Spec.Serviceaccount)
				log.Info("Creating a new scc rolebinding ", "service.Namespace", newrb.Namespace, "service.Name", newrb.Name)
				err = r.Create(ctx, newrb)
				if err != nil {
					log.Error(err, "Failed to create new scc rolebinding", "Service.Namespace", newrb.Namespace, "Service.Name", newrb.Name)
					return ctrl.Result{}, err
				}
				log.Info("sucessfuly created a new scc rolebinding", "service.Namespace", newrb.Namespace, "service.Name", newrb.Name)
			} /*else if err == nil {
				fmt.Printf(" the scc-rolebinding is already created ")
				//rb := &v1.RoleBinding{}
				//get the existing role binding and append the service account to that binding and return it

				//updaterb := r.RolebindingupdateForKasmmod(Kasmmod, rb)
				err = r.Update(ctx, updaterb)
				if err != nil {
					log.Error(err, "Failed to update scc rolebinding", "Service.Namespace", updaterb.Namespace, "Service.Name", updaterb.Name)
					return ctrl.Result{}, err
				}
				log.Info("sucessfuly updated the scc rolebinding", "service.Namespace", updaterb.Namespace, "service.Name", updaterb.Name)
			}*/
		} else if err != nil {
			log.Error(err, "Failed to get service account")
			log.Info("else if value", err)
		} /*else if Kasmmod.Spec.Serviceaccount != "" {
			log.Info("service account already exists")
		} */
	}

	// Check if the deployment already exists, if not create a new one
	found := &appsv1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{Name: Kasmmod.Name, Namespace: Kasmmod.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		// Define a new deployment
		dep := r.deploymentForKasmmod(Kasmmod)
		log.Info("Creating a new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
		err = r.Create(ctx, dep)
		if err != nil {
			log.Error(err, "Failed to create new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
			return ctrl.Result{}, err
		}
		log.Info("Succcessfully created new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
		// return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Deployment")
		return ctrl.Result{}, err
	}

	// Deployment created successfully - return and requeue
	//check if the service already exists, if not create a new service
	foundsvc := &corev1.Service{}
	err = r.Get(ctx, types.NamespacedName{Name: Kasmmod.Name, Namespace: Kasmmod.Namespace}, foundsvc)
	// fmt.Printf("\n this is the data of the service account on line 169 %s, %s\n", err, foundsvc)
	if err != nil && errors.IsNotFound(err) {
		// Define a new service
		svc := r.ServicecreationForKasmmod(Kasmmod)
		log.Info("Creating a new Service for Deployment", "Service.Namespace", svc.Namespace, "Service.Name", svc.Name)
		err = r.Create(ctx, svc)
		if err != nil {
			log.Error(err, "Failed to create new Service for the Deployment", "Service.Namespace", svc.Namespace, "Service.Name", svc.Name)
			return ctrl.Result{}, err
		}
		log.Info("Successfully created a new service for the deployment", "Service.Namespace", svc.Namespace, "Service.Name", svc.Name)
		// return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Service")
		return ctrl.Result{}, err
	}

	foundrt := &routev1.Route{}
	err = r.Get(ctx, types.NamespacedName{Name: Kasmmod.Name, Namespace: Kasmmod.Namespace}, foundrt)
	if err != nil && errors.IsNotFound(err) {
		// Define a new route for the service
		route := r.RoutecreationForKasmmod(Kasmmod)
		log.Info("Creating a new Route for Service", "Route.Namespace", route.Namespace, "Route.Name", route.Name)
		err = r.Create(ctx, route)
		if err != nil {
			log.Error(err, "Failed to create new Route for the Service", "Route.Namespace", route.Namespace, "Route.Name", route.Name)
			return ctrl.Result{}, err
		}
		log.Info("sucessfuly created a new Route for the Service", "Route.Namespace", route.Namespace, "Route.Name", route.Name)
		return ctrl.Result{}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Deployment")
		return ctrl.Result{}, err
	}

	// Ensure the deployment size is the same as the spec
	size := Kasmmod.Spec.Size //NewSize
	// *found.Spec.Replicas -> OldSize
	if *found.Spec.Replicas != size {
		found.Spec.Replicas = &size
		err = r.Update(ctx, found)
		if err != nil {
			log.Error(err, "Failed to update size of the Deployment", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)
			return ctrl.Result{}, err
		}
		log.Info("Successfully Updated the size of the Deployment", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)
		return ctrl.Result{}, nil
	}
	// Ensure the deployment port is the same as the spec
	port := Kasmmod.Spec.Port //NewPort
	//fmt.Printf("new-port: %v\n", port)
	//fmt.Printf("old-port: %v\n", found.Spec.Template.Spec.Containers[0].Ports[0].ContainerPort)
	// found.Spec.Template.Spec.Containers[0].Ports[0].ContainerPort -> OldPort
	if found.Spec.Template.Spec.Containers[0].Ports[0].ContainerPort != port && port != 0 {
		found.Spec.Template.Spec.Containers[0].Ports[0].ContainerPort = port
		err = r.Update(ctx, found)
		if err != nil {
			log.Error(err, "Failed to update port of the Deployment1", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)
			return ctrl.Result{}, err
		}
		log.Info("Successfully Updated the Port number of the Deployment", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)
		return ctrl.Result{}, nil
	}
	// Also Update the port in service of deployment
	oldport, err1 := GetPort(ctx, Kasmmod)
	//fmt.Printf("old-port: %v\n", oldport)
	if err1 != nil {
		log.Error(err1, "Failed to update Port number of Deployment2", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)
		return ctrl.Result{}, err1
	} else if oldport != port && port != 0 {
		err2 := UpdatePort(port, ctx, Kasmmod)
		if err2 != nil {
			log.Error(err2, "Failed to update Port number of Deployment3", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)
			return ctrl.Result{}, err2
		}
		log.Info("Successfully Updated the Port number in the Service", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)
		return ctrl.Result{}, nil
	}

	// Ensure the deployment image is the same as the spec
	image := Kasmmod.Spec.Image //NewImage
	// found.Spec.Template.Spec.Containers[0].Image -> OldImage
	//fmt.Printf("NewImage: %v\n", image)
	//fmt.Printf("OldImage: %v\n", found.Spec.Template.Spec.Containers[0].Image)
	if found.Spec.Template.Spec.Containers[0].Image != image && image != "" {
		found.Spec.Template.Spec.Containers[0].Image = image
		err = r.Update(ctx, found)
		if err != nil {
			log.Error(err, "Failed to update Image of the Deployment", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)
			return ctrl.Result{}, err
		}
		log.Info("Successfully Updated the image of the Deployment", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)
		return ctrl.Result{}, nil
	}
	// Ensure the deployment service account is same as the spec
	serviceAccountName := Kasmmod.Spec.Serviceaccount //NewServiceAccount
	/* check if the NewServiceAccount is present in the namespace,
	if not present then create a service account and rolebinding */
	foundsanew := &corev1.ServiceAccount{}
	errnew := r.Get(ctx, types.NamespacedName{Name: Kasmmod.Spec.Serviceaccount, Namespace: Kasmmod.Namespace}, foundsanew)
	//has not error which means there is a deployment already exist
	if errnew != nil && errors.IsNotFound(errnew) && Kasmmod.Spec.Serviceaccount != "" {
		sanew := r.ServiceaccountForKasmmod(Kasmmod, serviceAccountName)
		log.Info("Creating a new service account", "service.Namespace", sanew.Namespace, "service.Name", sanew.Name)
		errnew = r.Create(ctx, sanew)
		if errnew != nil {
			log.Error(errnew, "Failed to create new service account", "Service.Namespace", sanew.Namespace, "Service.Name", sanew.Name)
			return ctrl.Result{}, errnew
		}
		log.Info("sucessfuly created a new service account", "service.Namespace", sanew.Namespace, "service.Name", sanew.Name)
		//check if the rolebinding is avaiable
		/*This line initializes a new RoleBinding object using the v1 API version.
		It creates an empty RoleBinding instance that will be used to store the retrieved data.*/
		foundrbnew := &v1.RoleBinding{}
		errnew = r.Get(ctx, types.NamespacedName{Name: Kasmmod.Spec.Serviceaccount, Namespace: Kasmmod.Namespace}, foundrbnew)
		if errnew != nil && errors.IsNotFound(errnew) {
			rbnew := r.RolebindingcreationForKasmmod(Kasmmod, serviceAccountName)
			log.Info("Creating a new scc rolebinding ", "service.Namespace", rbnew.Namespace, "service.Name", rbnew.Name)
			errnew = r.Create(ctx, rbnew)
			if errnew != nil {
				log.Error(errnew, "Failed to create new scc rolebinding", "Service.Namespace", rbnew.Namespace, "Service.Name", rbnew.Name)
				return ctrl.Result{}, errnew
			}
			log.Info("sucessfuly created a new scc rolebinding", "service.Namespace", rbnew.Namespace, "service.Name", rbnew.Name)
		}
	}
	// found.Spec.Template.Spec.ServiceAccountName -> OldServiceAccount
	if found.Spec.Template.Spec.ServiceAccountName != serviceAccountName && serviceAccountName != "" {
		found.Spec.Template.Spec.ServiceAccountName = serviceAccountName
		err = r.Update(ctx, found)
		if err != nil {
			log.Error(err, "Failed to update ServiceAccount of Deployment", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)
			return ctrl.Result{}, err
		}
		log.Info("Successfully Updated the Service Account of the Deployment", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)
		return ctrl.Result{}, nil
	}
	//  Ensure the sessionid is same as that of spec
	sessionid := Kasmmod.Spec.Sessionid //NewSessionID
	// sessionid1 -> OldSessionID
	sessionid1, err1 := getSessionID(ctx, Kasmmod)
	//fmt.Printf("NewSessionID: %v\n", sessionid)
	//fmt.Printf("OldSessionID: %v\n", sessionid1)
	if err1 != nil {
		log.Error(err1, "Failed to update SessionID of Deployment1", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)
		return ctrl.Result{}, err1
	} else if sessionid1 != sessionid && sessionid != "" {
		err2 := UpdateSessionID(sessionid, ctx, Kasmmod)
		if err2 != nil {
			log.Error(err2, "Failed to update SessionID of Deployment2", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)
			return ctrl.Result{}, err2
		}
		log.Info("Successfully Updated the SessionID of the Deployment", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)
		return ctrl.Result{}, nil
	}

	// Update the Kasmmod status with the pod names
	// List the pods for this Kasmmod's deployment
	podList := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace(Kasmmod.Namespace),
		client.MatchingLabels(labelsForKasmmod(Kasmmod.Name)),
	}
	if err = r.List(ctx, podList, listOpts...); err != nil {
		log.Error(err, "Failed to list pods", "Kasmmod.Namespace", Kasmmod.Namespace, "Kasmmod.Name", Kasmmod.Name)
		return ctrl.Result{}, err
	}
	podNames := getPodNames(podList.Items)

	// Update status.Nodes if needed
	if !reflect.DeepEqual(podNames, Kasmmod.Status.Nodes) {
		Kasmmod.Status.Nodes = podNames
		Kasmmod.Status.State = "Deployment created successfully"
		err := r.Status().Update(ctx, Kasmmod)
		if err != nil {
			log.Error(err, "Failed to update Kasmmod status")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func getSessionID(ctx context.Context, m *ocpztpv1.Kasmmod) (string, error) {
	// Execute `oc` command to get the yaml of the route
	log := ctrllog.FromContext(ctx)
	routedata := exec.Command("oc", "get", "routes", m.Name, "-o", "yaml")
	output, err := routedata.Output()
	if err != nil {
		log.Error(err, "Couldn't acquire the session id for the kasmmod deployment")
		return "", err
	}
	//fmt.Printf("Route_data:\n%s", output)
	obj := make(map[string]interface{})
	err = yaml.Unmarshal(output, &obj)
	if err != nil {
		log.Error(err, "Failed to acquire sessionid for the deployment")
		return "", err
	}
	// Modify the fields in the struct as needed
	spec := obj["spec"].(map[interface{}]interface{})
	host := spec["host"].(string) //changing interface{} to string
	host1 := strings.TrimSpace(string(host))
	index := strings.Index(host1, ".")
	if index == -1 {
		return m.Name, nil
	}
	return host1[0:index], nil
}

func UpdateSessionID(NewSessionID string, ctx context.Context, m *ocpztpv1.Kasmmod) error {
	log := ctrllog.FromContext(ctx)
	routedata := exec.Command("oc", "get", "routes", m.Name, "-o", "yaml")
	output, err := routedata.Output()
	if err != nil {
		log.Error(err, "Couldn't acquire the session id for the kasmmod deployment")
		return err
		//return "",err
	}
	//fmt.Printf("Route_data:\n%s", output)
	obj := make(map[string]interface{})
	err = yaml.Unmarshal(output, &obj)
	if err != nil {
		log.Error(err, "Failed to acquire sessionid for the deployment")
		return err
		//return "",err
	}
	// Modify the fields in the struct as needed
	spec := obj["spec"].(map[interface{}]interface{})
	host := spec["host"].(string) //changing interface{} to string
	host1 := strings.TrimSpace(string(host))
	index := strings.Index(host1, ".")
	if index != -1 {
		newhost := NewSessionID + host1[index:]
		spec["host"] = newhost
	}
	// Convert the struct back to YAML
	modifiedOutput, err1 := yaml.Marshal(obj)
	if err1 != nil {
		log.Error(err1, "Failed to convert struct back to YAML")
		return err1
	}
	// print the modified Kasmmod deployment yaml
	//fmt.Printf("modified_route_data:\n%s\n", modifiedOutput)
	modifiedYAML := string(modifiedOutput)
	// Execute the kubectl apply command with the modified YAML content
	cmd := exec.Command("oc", "apply", "-f", "-")
	cmd.Stdin = bytes.NewBufferString(modifiedYAML)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err2 := cmd.Run()
	if err2 != nil {
		log.Error(err2, "SessionID update failed")
		return err2
	}
	//return host1[0:index],nil
	return nil
}

func GetPort(ctx context.Context, m *ocpztpv1.Kasmmod) (int32, error) {
	// Execute `oc` command to get the yaml of the route
	log := ctrllog.FromContext(ctx)
	servicedata := exec.Command("oc", "get", "service", m.Name, "-o", "yaml")
	output, err := servicedata.Output()
	if err != nil {
		log.Error(err, "Couldn't acquire the service for the kasmmod deployment")
		return 0, err
	}
	//fmt.Printf("Service_data:\n%s", output)
	obj := make(map[string]interface{})
	err = yaml.Unmarshal(output, &obj)
	if err != nil {
		log.Error(err, "Failed to acquire service for the deployment")
		return 0, err
	}
	// Update the port in the YAML
	spec := obj["spec"].(map[interface{}]interface{})
	ports := spec["ports"].([]interface{})
	portnum := ports[0].(map[interface{}]interface{})["port"].(int)
	portnum32 := int32(portnum)
	return portnum32, nil
}

func UpdatePort(NewPort int32, ctx context.Context, m *ocpztpv1.Kasmmod) error {
	// Execute `oc` command to get the yaml of the route
	log := ctrllog.FromContext(ctx)
	servicedata := exec.Command("oc", "get", "service", m.Name, "-o", "yaml")
	output, err := servicedata.Output()
	if err != nil {
		log.Error(err, "Couldn't acquire the service for the kasmmod deployment")
		return err
	}
	//fmt.Printf("Service_data:\n%s", output)
	obj := make(map[string]interface{})
	err = yaml.Unmarshal(output, &obj)
	if err != nil {
		log.Error(err, "Failed to acquire service for the deployment")
		return err
	}
	// Update the port in the YAML
	spec := obj["spec"].(map[interface{}]interface{})
	ports := spec["ports"].([]interface{})
	ports[0].(map[interface{}]interface{})["port"] = int(NewPort)
	ports[0].(map[interface{}]interface{})["targetPort"] = int(NewPort)
	// Convert the struct back to YAML
	modifiedOutput, err1 := yaml.Marshal(obj)
	if err1 != nil {
		log.Error(err1, "Failed to convert struct back to YAML")
		return err1
	}
	// print the modified Kasmmod deployment yaml
	//fmt.Printf("modified_service_data:\n%s\n", modifiedOutput)
	modifiedYAML := string(modifiedOutput)
	// Execute the kubectl apply command with the modified YAML content
	cmd := exec.Command("oc", "replace", "-f", "-")
	cmd.Stdin = bytes.NewBufferString(modifiedYAML)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err2 := cmd.Run()
	if err2 != nil {
		log.Error(err2, "Port Number update failed")
		return err2
	}
	return nil
}

func getClusterFQDN() string {
	// Execute the `oc` command to get the cluster's FQDN
	cmd := exec.Command("oc", "whoami", "--show-console")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	// Parse the output to extract the FQDN
	fqdn := strings.TrimSpace(string(output))
	index := strings.Index(fqdn, ".apps")
	if index == -1 {
		return fqdn
	}
	return fqdn[index+1:]
}

// deploymentForKasmmod returns a Kasmmod Deployment object
func (r *KasmmodReconciler) deploymentForKasmmod(m *ocpztpv1.Kasmmod) *appsv1.Deployment {
	Image := "balakuberox/python-web"
	if m.Spec.Image != "" {
		Image = m.Spec.Image
	}
	var Port int32 = 8080
	if m.Spec.Port != 0 {
		Port = m.Spec.Port
	}
	var serviceaccount string = "default"
	if m.Spec.Serviceaccount != "" {
		serviceaccount = m.Spec.Serviceaccount
	}
	//fmt.Printf("Serviceaccount:%s\n", serviceaccount)
	ls := labelsForKasmmod(m.Name)
	replicas := m.Spec.Size

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name,
			Namespace: m.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image: Image,
						Name:  "kasmmod",
						Ports: []corev1.ContainerPort{{
							ContainerPort: Port,
							Name:          "kasmmod",
						}},
					}},
					ServiceAccountName: serviceaccount,
				},
			},
		},
	}
	ctrl.SetControllerReference(m, dep, r.Scheme)
	return dep
}

// service account creation for the custom SCC functionality
func (r *KasmmodReconciler) ServiceaccountForKasmmod(m *ocpztpv1.Kasmmod, NewServiceAccountName string) *corev1.ServiceAccount {
	ls := labelsForKasmmod(m.Name)
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Labels:    ls,
			Name:      NewServiceAccountName,
			Namespace: m.Namespace,
		},
	}
	ctrl.SetControllerReference(m, sa, r.Scheme)
	return sa
}

// rolebinding creation using the CR
func (r *KasmmodReconciler) RolebindingcreationForKasmmod(m *ocpztpv1.Kasmmod, NewServiceAccountName string) *v1.RoleBinding {
	ls := labelsForKasmmod(m.Name)
	rb := &v1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      NewServiceAccountName + "-kasm-scc",
			Labels:    ls,
			Namespace: m.Namespace,
		},
		Subjects: []v1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      m.Spec.Serviceaccount,
				Namespace: m.Namespace,
			},
		},
		RoleRef: v1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     "system:openshift:scc:final-scc",
		},
	}
	ctrl.SetControllerReference(m, rb, r.Scheme)
	return rb
	//system:openshift:scc:final-scc
}

// service creation using the CR
func (r *KasmmodReconciler) ServicecreationForKasmmod(m *ocpztpv1.Kasmmod) *corev1.Service {
	ls := labelsForKasmmod(m.Name)
	var port int32
	port = 8080
	if m.Spec.Port != 0 {
		port = m.Spec.Port
	}
	portstr := intstr.FromInt(int(port))
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: m.Name,
			//Labels:    ls,
			Namespace: m.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: ls,
			Ports: []corev1.ServicePort{
				{
					//Name:       m.Name + "-port-1",
					Port:       port,
					TargetPort: portstr,
				},
			},
		},
	}
	ctrl.SetControllerReference(m, service, r.Scheme)
	return service
}

func (r *KasmmodReconciler) RoutecreationForKasmmod(m *ocpztpv1.Kasmmod) *routev1.Route {
	routeinfo := getClusterFQDN()
	session := m.Name
	if m.Spec.Sessionid != "" {
		session = m.Spec.Sessionid
	}
	route := &routev1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name,
			Namespace: m.Namespace,
		},
		Spec: routev1.RouteSpec{
			Host: session + "." + routeinfo,
			To: routev1.RouteTargetReference{
				Kind: "Service",
				Name: m.Name,
			},
			TLS: &routev1.TLSConfig{
				Termination: routev1.TLSTerminationEdge,
			},
			WildcardPolicy: "None",
		},
	}
	ctrl.SetControllerReference(m, route, r.Scheme)
	return route
}

// belonging to the given Kasmmod CR name.
func labelsForKasmmod(name string) map[string]string {
	return map[string]string{"app": "Kasmmod", "Kasmmod_cr": name}
}

// getPodNames returns the pod names of the array of pods passed in
func getPodNames(pods []corev1.Pod) []string {
	var podNames []string
	for _, pod := range pods {
		podNames = append(podNames, pod.Name)
	}
	return podNames
}

// SetupWithManager sets up the controller with the Manager.
func (r *KasmmodReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&ocpztpv1.Kasmmod{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}
