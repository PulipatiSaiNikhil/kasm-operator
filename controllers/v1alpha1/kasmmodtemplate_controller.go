/*
Copyright 2023.

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

	ocpztpv1alpha1 "github.com/balakuberox/kasmmod.git/api/v1alpha1"
	"gopkg.in/yaml.v2"

	//"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"

	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

// KasmmodTemplateReconciler reconciles a KasmmodTemplate object
type KasmmodTemplateReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=ocpztp.kasmmod.office.ocpztp.com,resources=kasmmodtemplates,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=ocpztp.kasmmod.office.ocpztp.com,resources=kasmmodtemplates/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=ocpztp.kasmmod.office.ocpztp.com,resources=kasmmodtemplates/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the KasmmodTemplate object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *KasmmodTemplateReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrllog.FromContext(ctx)
	// Fetch the KasmmodTemplate instance
	KasmmodTemplate := &ocpztpv1alpha1.KasmmodTemplate{}
	//Kasmmod := &ocpztpv1.Kasmmod{}
	err := r.Get(ctx, req.NamespacedName, KasmmodTemplate)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("KasmmodTemplate resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get KasmmodTemplate")
		return ctrl.Result{}, err
	}
	found1 := &appsv1.Deployment{}
	err = r.Get(context.TODO(), types.NamespacedName{Name: KasmmodTemplate.Name, Namespace: KasmmodTemplate.Namespace}, found1)
	if err != nil && errors.IsNotFound(err) {
		// if the deployment of KasmmodTemplate doesnot exists
		kasmmoddata := exec.Command("oc", "get", "kasmmod", KasmmodTemplate.Spec.Targetkasmmod, "-n", KasmmodTemplate.Namespace, "-o", "yaml")
		output, err := kasmmoddata.Output()
		if err != nil {
			log.Error(err, "Failed to update the existing Deployment because of error in output of the command", "Deployment.Namespace", KasmmodTemplate.Namespace, "Deployment.Name", KasmmodTemplate.Spec.Targetkasmmod)
			return ctrl.Result{}, err
		} else {
			//print the output of the command
			//fmt.Printf("kasmmod_data:\n%s", output)
			obj := make(map[string]interface{})
			err := yaml.Unmarshal(output, &obj)
			if err != nil {
				log.Error(err, "Failed to Update existing Deployment template")
				return ctrl.Result{}, err
			}
			// Modify the fields in the struct as needed
			spec := obj["spec"].(map[interface{}]interface{})
			var replicas int32 = KasmmodTemplate.Spec.Size
			if replicas != 0 {
				spec["size"] = replicas
			}
			if KasmmodTemplate.Spec.Image != "" {
				spec["image"] = KasmmodTemplate.Spec.Image
			}
			var port int32 = KasmmodTemplate.Spec.Port
			if port != 0 {
				spec["port"] = port
			}
			if KasmmodTemplate.Spec.Serviceaccount != "" {
				spec["serviceaccount"] = KasmmodTemplate.Spec.Serviceaccount
			}
			//fmt.Printf("ServiceAccount: %s\n", spec["serviceaccount"])
			if KasmmodTemplate.Spec.Sessionid != "" {
				spec["sessionid"] = KasmmodTemplate.Spec.Sessionid
			}
			// Convert the struct back to YAML
			modifiedOutput, err := yaml.Marshal(obj)
			if err != nil {
				log.Error(err, "Kasmmod deployment can't be updated")
				return ctrl.Result{}, err
			}
			// print the modified Kasmmod deployment yaml
			//fmt.Printf("modified_data:\n%s\n", modifiedOutput)
			// Convert the modified output to a string
			modifiedYAML := string(modifiedOutput)
			// Execute the kubectl apply command with the modified YAML content
			cmd := exec.Command("oc", "apply", "-f", "-")
			cmd.Stdin = bytes.NewBufferString(modifiedYAML)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			err = cmd.Run()
			if err != nil {
				log.Error(err, "Kasmmod Deployment update failed")
				return ctrl.Result{}, err
			}
			log.Info("Kasmmod Deployment Updated successfully", "Deployment.Namespace", KasmmodTemplate.Namespace, "Deployment.Name", KasmmodTemplate.Spec.Targetkasmmod)
		}
	} else {
		// deployment of kasmmod template exists
		log.Error(err, "Failed to update the Deployment")
		return ctrl.Result{}, err
	}

	/*found := &ocpztpv1.Kasmmod{}
	//log.Info("Kasmmod Info", "found.Spec.Image", found.Spec.Image, "found.Spec.Serviceaccount", found.Spec.Serviceaccount)
	err = r.Get(context.TODO(), types.NamespacedName{Name: KasmmodTemplate.Spec.Targetkasmmod, Namespace: KasmmodTemplate.Namespace}, found)
	fmt.Printf("error:%d\n", err)
	if err != nil && errors.IsAlreadyExists(err) {
		// the Kasmmod CR already exists and we have to update it with the new values from the KasmmodTemplate CR
		dep := r.updateKasmmod(found, KasmmodTemplate)
		log.Info("Updating the Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
		err = r.Update(ctx, dep)
		if err != nil {
			log.Error(err, "Failed to update the existing Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
			return ctrl.Result{}, err
		}
		log.Info("Updated successfully", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
	} else if err != nil && errors.IsNotFound(err) {
		// the CR doesnt exists and we have to create a new deployment of kind Kasmmod from
		//the KasmmodTemplate CR with a name of KasmmodTemplate.Spec.TargetKasmmod
		dep := r.updateKasmmod(found, KasmmodTemplate)
		log.Info("Updating the Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
		err = r.Update(ctx, dep)
		if err != nil {
			log.Error(err, "Failed to update the existing Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
			return ctrl.Result{}, err
		}
		log.Info("Deployment Created using KasmmodTemplate Spec!", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to update/create Deployment")
		return ctrl.Result{}, err
	} else {
		log.Info("not updating")
	}*/

	// Deployment created successfully - return and requeue

	// Update the Kasmmod status with the pod names
	// List the pods for this Kasmmod's deployment
	podList := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace(KasmmodTemplate.Namespace),
		client.MatchingLabels(labelsForKasmmodTemplate(KasmmodTemplate.Name)),
	}
	if err = r.List(ctx, podList, listOpts...); err != nil {
		log.Error(err, "Failed to list pods", "Kasmmod.Namespace", KasmmodTemplate.Namespace, "Kasmmod.Name", KasmmodTemplate.Name)
		return ctrl.Result{}, err
	}
	podNames := getPodNames(podList.Items)

	// Update status.Nodes if needed
	if !reflect.DeepEqual(podNames, KasmmodTemplate.Status.Nodes) {
		KasmmodTemplate.Status.Nodes = podNames
		KasmmodTemplate.Status.State = "Deployment created successfully"
		err := r.Status().Update(ctx, KasmmodTemplate)
		if err != nil {
			log.Error(err, "Failed to update Kasmmod status")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// deploymentForKasmmod returns a Kasmmod Deployment object

/*func (r *KasmmodTemplateReconciler) updateKasmmod(m1 *ocpztpv1.Kasmmod, m *ocpztpv1alpha1.KasmmodTemplate) *ocpztpv1.Kasmmod {
	var replicas int32 = m.Spec.Size
	if replicas != 0 {
		m1.Spec.Port = replicas
	}
	if m.Spec.Image != "" {
		m1.Spec.Image = m.Spec.Image
	}
	var port int32 = m.Spec.Port
	if port != 0 {
		m1.Spec.Port = port
	}
	if m.Spec.Serviceaccount != "" {
		m1.Spec.Serviceaccount = m.Spec.Serviceaccount
	}
	if m.Spec.Sessionid != "" {
		m1.Spec.Sessionid = m.Spec.Sessionid
	}
	Image := "balakuberox/python-web"
	if m.Spec.Image != "" {
		Image = m.Spec.Image
	}

	//ls := labelsForKasmmodTemplate(m.Name)
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
						Name:  "kasmmodtemplate",
					}},
				},
			},
		},
	}
	ctrl.SetControllerReference(m, dep, r.Scheme)
	return dep
	return m1
}*/

// belonging to the given KasmmodTemplate CR name.
func labelsForKasmmodTemplate(name string) map[string]string {
	return map[string]string{"app": "KasmmodTemplate", "KasmmodTemplate_cr": name}
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
func (r *KasmmodTemplateReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&ocpztpv1alpha1.KasmmodTemplate{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}
