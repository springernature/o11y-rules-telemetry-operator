/*
Copyright 2024.

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

// https://sklar.rocks/kubernetes-custom-resource-definitions/
// https://github.com/slaise/community-operators/blob/master/docs/best-practices.md
// https://stuartleeks.com/posts/kubebuilder-event-filters-part-2-update/
// https://janosmiko.com/blog/2023-03-04-tutorial-kubebuilder-3/
// https://itnext.io/kubernetes-custom-controllers-recipes-for-beginners-bbc286c05ef8

package controller

import (
	"context"
	"fmt"
	"net/url"
	"reflect"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	record "k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	builder "sigs.k8s.io/controller-runtime/pkg/builder"
	client "sigs.k8s.io/controller-runtime/pkg/client"
	controller "sigs.k8s.io/controller-runtime/pkg/controller"
	controllerutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	event "sigs.k8s.io/controller-runtime/pkg/event"
	log "sigs.k8s.io/controller-runtime/pkg/log"
	predicate "sigs.k8s.io/controller-runtime/pkg/predicate"

	mimirrulesv1beta1 "springernature/o11y-rules-telemetry-operator/api/v1beta1"
)

const (
	EnvMimirRulerAPI              = "CONTROLLER_MIMIR_API"
	EnvMimirRulerCRDLabelSelector = "CONTROLLER_CR_SELECTOR"
	EnvMimirRulerTenantAnnotation = "CONTROLLER_CR_MIMIR_TENANT_ANNOTATION"
	MimirRulerTenantAnnotation    = "telemetry.springernature.com/o11y-tenant"
	MimirRulerControllerFinalizer = "telemetry.springernature.com/mimirrules-finalizer"
)

// MimirRulesReconciler reconciles a MimirRules object
type MimirRulesReconciler struct {
	client.Client
	Scheme             *runtime.Scheme
	mimirRulerClient   *MimirRulerClient
	Recorder           record.EventRecorder
	MimirApi           *url.URL
	CRTenantAnnotation string
	CRLabelSelector    *metav1.LabelSelector
}

//+kubebuilder:rbac:groups=mimirrules.telemetry.springernature.com,resources=mimirrules,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=mimirrules.telemetry.springernature.com,resources=mimirrules/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=mimirrules.telemetry.springernature.com,resources=mimirrules/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.0/pkg/reconcile
func (r *MimirRulesReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// Add a deadline just to make sure we don't get stuck in a loop
	ctx, cancel := context.WithDeadline(ctx, metav1.Now().Add(60*time.Second))
	defer cancel()
	// set the logger with the context
	log := log.FromContext(ctx)
	// Get the MimirRules resource that triggered the reconciliation request
	var mimirRules mimirrulesv1beta1.MimirRules
	if err := r.Get(ctx, req.NamespacedName, &mimirRules); err != nil {
		log.Error(err, "Unable to fetch MimirRules object")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// get the name of object
	name := mimirRules.GetName()
	// set the tenant from the namespace ...
	desiredTenant := mimirRules.GetNamespace()
	if value, found := mimirRules.GetAnnotations()[r.CRTenantAnnotation]; found {
		// ... or from the annotation if it is present
		desiredTenant = value
	}
	log.Info("Reconciling MimirRules", "tenant", desiredTenant)

	if mimirRules.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object. This is equivalent
		// to registering our finalizer.
		if !controllerutil.ContainsFinalizer(&mimirRules, MimirRulerControllerFinalizer) {
			controllerutil.AddFinalizer(&mimirRules, MimirRulerControllerFinalizer)
			if err := r.Update(ctx, &mimirRules); err != nil {
				log.Error(err, "Error adding finalizer")
				return ctrl.Result{}, err
			}
		}
		// Add all ruleGroup to Mimir namespace "name"
		status := make(map[string]string)
		errors := 0
		for _, group := range mimirRules.Spec.Groups {
			data, _ := group.Rules.MarshalJSON()
			rules, err := NewMimirGroupRules(group.GroupName, data)
			if err != nil {
				status[group.GroupName] = err.Error()
				log.Error(err, "Error parsing RuleGroup", "groupName", group.GroupName)
				r.Recorder.Event(&mimirRules, corev1.EventTypeWarning, "update-error", err.Error())
				errors++
			} else {
				if msg, err := r.mimirRulerClient.SetGroupRules(desiredTenant, name, rules); err != nil {
					status[group.GroupName] = err.Error()
					log.Error(err, "Error setting RuleGroup", "groupName", group.GroupName)
					r.Recorder.Event(&mimirRules, corev1.EventTypeWarning, "update-error", err.Error())
					errors++
				} else {
					status[group.GroupName] = msg
					log.Info("Added RuleGroup", "groupName", group.GroupName)
					r.Recorder.Eventf(&mimirRules, corev1.EventTypeNormal, "update", "group %s on namespace %s", group.GroupName, name)
				}
			}
		}
		// Delete GroupRules which are not present anymore in the current object
		if currentRules, err := r.mimirRulerClient.GetGroupRules(desiredTenant, name); err == nil {
			for _, currentGroup := range currentRules {
				found := false
				for _, group := range mimirRules.Spec.Groups {
					if currentGroup.GroupName == group.GroupName {
						found = true
						break
					}
				}
				if !found {
					if _, err := r.mimirRulerClient.DeleteGroupRules(desiredTenant, name, currentGroup.GroupName); err != nil {
						log.Error(err, "Error deleting RuleGroup", "groupName", currentGroup.GroupName)
						r.Recorder.Event(&mimirRules, corev1.EventTypeWarning, "delete-error", err.Error())
					} else {
						log.Info("Deleted RuleGroup", "groupName", currentGroup.GroupName)
						r.Recorder.Eventf(&mimirRules, corev1.EventTypeNormal, "delete", "group %s on namespace %s", currentGroup.GroupName, name)
					}

				}
			}
		} else {
			log.Error(err, "Error getting current RuleGroups to delete undefined")
			r.Recorder.Event(&mimirRules, corev1.EventTypeWarning, "delete-error", err.Error())
		}
		// when a user move rules to a different tenant (eg. by adding/changing annotation)
		currentTenant := mimirRules.Status.Tenant
		if currentTenant == "" {
			// this is a new object, it has no status yet
			currentTenant = desiredTenant
		}
		if desiredTenant != currentTenant {
			_, err := r.mimirRulerClient.DeleteNamespace(currentTenant, name)
			if err != nil {
				log.Error(err, "Error deleting Ruler Namespace from old tenant", "oldtenant", currentTenant, "newtenant", desiredTenant)
				r.Recorder.Event(&mimirRules, corev1.EventTypeWarning, "move-error", err.Error())
			} else {
				log.Info("Deleted Ruler Namespace from old tenant", "oldtenant", currentTenant, "newtenant", desiredTenant)
				r.Recorder.Eventf(&mimirRules, corev1.EventTypeNormal, "move", "reassign owner from %s to %s", currentTenant, desiredTenant)
			}
		}
		// Update status
		mimirRules.Status.GroupStatus = status
		mimirRules.Status.Errors = errors
		mimirRules.Status.Tenant = desiredTenant
		mimirRules.Status.LastUpdate = time.Now().Format(time.RFC3339)
		if err := r.Status().Update(ctx, &mimirRules); err != nil {
			log.Error(err, "Unable to update status", "status", status)
			return ctrl.Result{}, err
		}
	} else {
		// The object is being deleted
		if controllerutil.ContainsFinalizer(&mimirRules, MimirRulerControllerFinalizer) {
			// Delete rules from Mimir
			_, err := r.mimirRulerClient.DeleteNamespace(desiredTenant, name)
			if err != nil {
				log.Error(err, "Error deleting Ruler Namespace", "tenant", desiredTenant)
				r.Recorder.Event(&mimirRules, corev1.EventTypeWarning, "delete-error", err.Error())
			} else {
				// remove our finalizer from the list and update it.
				controllerutil.RemoveFinalizer(&mimirRules, MimirRulerControllerFinalizer)
				if err := r.Update(ctx, &mimirRules); err != nil {
					log.Error(err, "Error removing finalizer")
					return ctrl.Result{}, err
				}
				log.Info("Deleted Ruler Namespace from tenant", "tenant", desiredTenant)
				r.Recorder.Eventf(&mimirRules, corev1.EventTypeNormal, "delete", "delete namespace from tenant %s", desiredTenant)
			}
		}
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MimirRulesReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.mimirRulerClient = NewMimirRulerClient(r.MimirApi)
	selectorFilter, err := predicate.LabelSelectorPredicate(*r.CRLabelSelector)
	if err != nil {
		return fmt.Errorf("unable to create predicate for selecting CRs: %s", err.Error())
	}
	// Filter reconcile events
	eventsFilters := predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			// Generation is only updated on spec changes (also on deletion), not metadata or status
			// Filter out events where the generation hasn't changed to avoid being triggered by status updates
			if e.ObjectOld == nil || e.ObjectNew == nil {
				return false
			}
			genChange := e.ObjectNew.GetGeneration() != e.ObjectOld.GetGeneration()
			labelsChange := !reflect.DeepEqual(e.ObjectNew.GetLabels(), e.ObjectOld.GetLabels())
			annotationsChange := !reflect.DeepEqual(e.ObjectNew.GetAnnotations(), e.ObjectOld.GetAnnotations())
			// gen, labels or annotations update
			return genChange || labelsChange || annotationsChange
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			// The reconciler adds a finalizer so we perform clean-up
			// when the delete timestamp is added
			// Suppress Delete events to avoid filtering them out in the Reconcile function
			return false
		},
	}
	// Create controller
	return ctrl.NewControllerManagedBy(mgr).
		Named("mimirrules_controller").
		// by default, an operator will only run a single reconcile loop per-controller.
		// when running a globally-scoped controller, it's useful to run multiple concurrent
		// reconcile loops to simultaneously handle many resource changes at once.
		WithOptions(controller.Options{MaxConcurrentReconciles: 3}).
		For(&mimirrulesv1beta1.MimirRules{}, builder.WithPredicates(selectorFilter)).
		WithEventFilter(eventsFilters).
		Complete(r)
}
