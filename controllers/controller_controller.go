/*


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
	"context"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	batchv1 "github.com/eilinge/replica-operator/api/v1"
	"github.com/eilinge/replica-operator/util"
	appsv1 "k8s.io/api/apps/v1"
)

var (
	ownerKey = ".metadata.controller"
	apiGVStr = batchv1.GroupVersion.String()
)

// ControllerReconciler reconciles a Controller object
type ControllerReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=batch.controller.kubebuilder.io,resources=controllers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=batch.controller.kubebuilder.io,resources=controllers/status,verbs=get;update;patch

func (r *ControllerReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	_ = r.Log.WithValues("controller", req.NamespacedName)
	var (
		name      string
		namespace string
		cou       int64
	)

	// your logic here
	contr := &batchv1.Controller{}
	// get replicas name and counts
	if err := r.Get(ctx, req.NamespacedName, contr); err != nil {
		klog.Error("unable to fetch contr", err)
	} else {
		name = contr.Spec.Name
		namespace = contr.Spec.Namespace
		cou = contr.Spec.Count
	}

	// Install RBAC resources for the Controller plugin kubernetes
	cr, sa, crb := util.MakeRBACObjects(contr.Name, contr.Namespace)
	// Set ServiceAccount's owner to this Controller
	if err := ctrl.SetControllerReference(contr, &sa, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}
	if err := r.Create(ctx, &cr); err != nil && !errors.IsAlreadyExists(err) {
		return ctrl.Result{}, err
	}
	if err := r.Create(ctx, &sa); err != nil && !errors.IsAlreadyExists(err) {
		return ctrl.Result{}, err
	}
	if err := r.Create(ctx, &crb); err != nil && !errors.IsAlreadyExists(err) {
		return ctrl.Result{}, err
	}

	// get deployment and update replicas
	if err := r.SetOwnerDeployment(name, namespace, contr); err != nil {
		klog.Error("Set Owner Deployment failed", err)
	}
	count := int32(cou)
	// deploy := de.DeepCopy() // deepcopy due to update replicas change failed
	if _, err := r.GetAndUpdateDeployment(name, namespace, &count); err != nil {
		klog.Error("update deployment failed", err)
	}

	return ctrl.Result{}, nil
}

// set Index for searching refer deployment
func (r *ControllerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(&appsv1.Deployment{}, ownerKey, func(rawObj runtime.Object) []string {
		// grab the deploy object, extract the owner.
		dm := rawObj.(*appsv1.Deployment)
		owner := metav1.GetControllerOf(dm)
		if owner == nil {
			return nil
		}
		// Make sure it's a Controller. If so, return it.
		if owner.APIVersion != apiGVStr || owner.Kind != "Controller" {
			return nil
		}
		return []string{owner.Name}
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&batchv1.Controller{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}

func (r *ControllerReconciler) GetAndUpdateDeployment(name, namespace string, count *int32) (deploy *appsv1.Deployment, err error) {
	dm := &appsv1.Deployment{}
	ctx := context.Background()
	err = r.Get(ctx, client.ObjectKey{Namespace: namespace, Name: name}, dm)
	if err == nil {
		if *dm.Spec.Replicas != *count {
			dm.Spec.Replicas = count
			if err := r.Update(ctx, dm); err != nil && !errors.IsNotFound(err) {
				return nil, err
			}
		}
	} else if errors.IsNotFound(err) {
		return nil, err
	}
	return dm, nil
}

func (r *ControllerReconciler) SetOwnerDeployment(name, namespace string, contr *batchv1.Controller) (err error) {
	dm := &appsv1.Deployment{}
	ctx := context.Background()
	err = r.Get(ctx, client.ObjectKey{Namespace: namespace, Name: name}, dm)
	if err == nil {
		var referexist bool
		for _, refer := range dm.ObjectMeta.OwnerReferences {
			if refer.APIVersion == apiGVStr && refer.Kind == "Controller" {
				referexist = true
			}
		}
		if !referexist {
			ownerKey := schema.GroupVersionKind{Kind: "Controller", Version: apiGVStr}
			references := []metav1.OwnerReference{*metav1.NewControllerRef(contr, ownerKey)}
			dm.SetOwnerReferences(references)
			if err := r.Update(ctx, dm); err != nil && !errors.IsNotFound(err) {
				return err
			}
		}
	} else if errors.IsNotFound(err) {
		return err
	}
	return nil
}
