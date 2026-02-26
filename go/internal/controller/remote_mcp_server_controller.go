/*
Copyright 2025.

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
	"context"
	"time"

	"github.com/kagent-dev/kagent/go/api/v1alpha2"
	"github.com/kagent-dev/kagent/go/internal/controller/reconciler"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// RemoteMCPServerController reconciles a RemoteMCPServer object
type RemoteMCPServerController struct {
	Scheme     *runtime.Scheme
	Reconciler reconciler.KagentReconciler
}

// +kubebuilder:rbac:groups=kagent.dev,resources=remotemcpservers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=kagent.dev,resources=remotemcpservers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=kagent.dev,resources=remotemcpservers/finalizers,verbs=update

func (r *RemoteMCPServerController) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	err := r.Reconciler.ReconcileKagentRemoteMCPServer(ctx, req)
	if err != nil {
		// Return zero result when there's an error - controller-runtime will handle backoff
		return ctrl.Result{}, err
	}
	// Success - requeue after 60s to refresh tool server status
	return ctrl.Result{
		RequeueAfter: 60 * time.Second,
	}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *RemoteMCPServerController) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		WithOptions(controller.Options{
			NeedLeaderElection: ptr.To(true),
		}).
		For(&v1alpha2.RemoteMCPServer{}).
		// Watch Secrets that are referenced in RemoteMCPServer.Spec.HeadersFrom so that
		// token rotations trigger an immediate re-reconciliation (tool discovery with the
		// new token) rather than waiting for the periodic 60-second requeue.
		Watches(
			&corev1.Secret{},
			handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
				return r.findRemoteMCPServersUsingSecret(ctx, mgr.GetClient(), types.NamespacedName{
					Name:      obj.GetName(),
					Namespace: obj.GetNamespace(),
				})
			}),
			builder.WithPredicates(predicate.ResourceVersionChangedPredicate{}),
		).
		// Watch ConfigMaps referenced in RemoteMCPServer.Spec.HeadersFrom for the same reason.
		Watches(
			&corev1.ConfigMap{},
			handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
				return r.findRemoteMCPServersUsingConfigMap(ctx, mgr.GetClient(), types.NamespacedName{
					Name:      obj.GetName(),
					Namespace: obj.GetNamespace(),
				})
			}),
			builder.WithPredicates(predicate.ResourceVersionChangedPredicate{}),
		).
		Named("remotemcpserver").
		Complete(r)
}

// findRemoteMCPServersUsingSecret returns reconcile requests for every RemoteMCPServer
// that references the given Secret in its HeadersFrom list.
func (r *RemoteMCPServerController) findRemoteMCPServersUsingSecret(ctx context.Context, cl client.Client, secretRef types.NamespacedName) []reconcile.Request {
	return r.findRemoteMCPServersUsingSource(ctx, cl, v1alpha2.SecretValueSource, secretRef)
}

// findRemoteMCPServersUsingConfigMap returns reconcile requests for every RemoteMCPServer
// that references the given ConfigMap in its HeadersFrom list.
func (r *RemoteMCPServerController) findRemoteMCPServersUsingConfigMap(ctx context.Context, cl client.Client, configMapRef types.NamespacedName) []reconcile.Request {
	return r.findRemoteMCPServersUsingSource(ctx, cl, v1alpha2.ConfigMapValueSource, configMapRef)
}

func (r *RemoteMCPServerController) findRemoteMCPServersUsingSource(ctx context.Context, cl client.Client, sourceType v1alpha2.ValueSourceType, ref types.NamespacedName) []reconcile.Request {
	var serverList v1alpha2.RemoteMCPServerList
	if err := cl.List(ctx, &serverList); err != nil {
		log.FromContext(ctx).Error(err, "failed to list RemoteMCPServers for Secret/ConfigMap watch")
		return nil
	}

	var requests []reconcile.Request
	for _, server := range serverList.Items {
		for _, h := range server.Spec.HeadersFrom {
			if h.ValueFrom == nil {
				continue
			}
			if h.ValueFrom.Type != sourceType {
				continue
			}
			sourceRef := types.NamespacedName{Namespace: server.Namespace, Name: h.ValueFrom.Name}
			if sourceRef == ref {
				requests = append(requests, reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      server.Name,
						Namespace: server.Namespace,
					},
				})
				break // one match per server is enough
			}
		}
	}
	return requests
}
