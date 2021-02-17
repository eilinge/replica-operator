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
	"sync"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	k8sinformers "k8s.io/client-go/informers"
	"k8s.io/klog"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	batchv1 "batch.controller.kubebuilder.io/replica/api/v1"
	v1 "batch.controller.kubebuilder.io/replica/api/v1"
	appsv1 "k8s.io/api/apps/v1"
)

var (
	k8sOnce         sync.Once
	informerFactory k8sinformers.SharedInformerFactory
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
	ticker := time.NewTicker(time.Second * 20)
	defer ticker.Stop()

RECONCILE:
	for {
		select {
		case <-ticker.C:
			contr := &v1.Controller{}
			// get replicas name and counts
			if err := r.Get(ctx, req.NamespacedName, contr); err != nil {
				klog.Error("unable to fetch contr", err)
				break RECONCILE
			} else {
				name = contr.Spec.Name
				namespace = contr.Spec.Namespace
				cou = contr.Spec.Count
			}
			// get deployment and update replicas
			count := int32(cou)
			// deploy := de.DeepCopy() // deepcopy due to update replicas change failed
			if _, err := r.GetAndUpdateDeployment(name, namespace, &count); err != nil {
				klog.Error("update deployment failed", err)
				break RECONCILE
			}
		}
	}

	return ctrl.Result{}, nil
}

func (r *ControllerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&batchv1.Controller{}).
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
