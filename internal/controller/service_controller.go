/*
Copyright 2023 GitHub, Inc.

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

package controller

import (

	// Stdlib
	"context"
	"fmt"

	// Community
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// ServiceReconciler reconciles a Service object
type ServiceReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	CommChan chan<- []string
}

const (
	controllerName = "swarm"
	appLabel       = "swarm"
)

//-----------------------------------------------------------------------------
// SetupWithManager sets up the controller with the Manager.
//-----------------------------------------------------------------------------

func (r *ServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {

	// Define the label selector as a predicate
	labelPredicate := predicate.NewPredicateFuncs(func(obj client.Object) bool {
		return obj.GetLabels()["app"] == appLabel
	})

	// Create the controller
	return ctrl.NewControllerManagedBy(mgr).
		Named(controllerName).
		For(&corev1.Service{}).
		WithEventFilter(labelPredicate).
		Complete(r)
}

//+kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=services/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=core,resources=services/finalizers,verbs=update

//-----------------------------------------------------------------------------
// Reconcile is part of the main kubernetes reconciliation loop.
//-----------------------------------------------------------------------------

func (r *ServiceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	// Set up logging
	log := log.Log.WithName(controllerName).WithValues("service", req.Name)

	// Get all the swarm services
	var services corev1.ServiceList
	if err := r.List(ctx, &services, client.MatchingLabels{"app": appLabel}); err != nil {
		log.Error(err, "unable to list services")
		return ctrl.Result{}, err
	}

	// Log this reconciliation
	log.V(1).Info("reconcile")

	// Send the services to the comm channel
	var serviceNames []string
	for _, service := range services.Items {
		for _, port := range service.Spec.Ports {
			if port.Name == "http" {
				serviceNames = append(serviceNames, service.Name+"."+service.Namespace+":"+fmt.Sprint(port.Port))
			}
		}
	}
	r.CommChan <- serviceNames

	// Return on success
	return ctrl.Result{}, nil
}
