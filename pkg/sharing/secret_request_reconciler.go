// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package sharing

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	sgv1alpha1 "github.com/vmware-tanzu/carvel-secretgen-controller/pkg/apis/secretgen/v1alpha1"
	"github.com/vmware-tanzu/carvel-secretgen-controller/pkg/reconciler"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	// metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// SecretRequestReconciler creates an imported Secret if it was exported.
type SecretRequestReconciler struct {
	client        client.Client
	secretExports *SecretExports
	log           logr.Logger
}

var _ reconcile.Reconciler = &SecretRequestReconciler{}

// NewSecretRequestReconciler constructs SecretRequestReconciler.
func NewSecretRequestReconciler(client client.Client,
	secretExports *SecretExports, log logr.Logger) *SecretRequestReconciler {
	return &SecretRequestReconciler{client, secretExports, log}
}

func (r *SecretRequestReconciler) AttachWatches(controller controller.Controller) error {
	err := controller.Watch(&source.Kind{Type: &sgv1alpha1.SecretRequest{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return fmt.Errorf("Watching secret request: %s", err)
	}

	// Watch secrets and enqueue for same named SecretRequest
	// to make sure imported secret is up-to-date
	err = controller.Watch(&source.Kind{Type: &corev1.Secret{}}, handler.EnqueueRequestsFromMapFunc(
		func(a client.Object) []reconcile.Request {
			return []reconcile.Request{
				{NamespacedName: types.NamespacedName{
					Name:      a.GetName(),
					Namespace: a.GetNamespace(),
				}},
			}
		},
	))
	if err != nil {
		return err
	}

	// Watch SecretExport and enqueue for related SecretRequest
	// based on export namespace configuration
	return controller.Watch(&source.Kind{Type: &sgv1alpha1.SecretExport{}}, &enqueueSecretExportToSecret{
		SecretExports: r.secretExports,
		Log:           r.log,

		ToRequests: func(_ client.Object) []reconcile.Request {
			var secretReqList sgv1alpha1.SecretRequestList

			// TODO expensive call on every secret export update
			err := r.client.List(context.TODO(), &secretReqList)
			if err != nil {
				// TODO what should we really do here?
				r.log.Error(err, "Failed fetching list of all secret requests")
				return nil
			}

			var result []reconcile.Request
			for _, req := range secretReqList.Items {
				result = append(result, reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      req.Name,
						Namespace: req.Namespace,
					},
				})
			}

			r.log.Info("Planning to reconcile matched secret requests",
				"all", len(secretReqList.Items))

			return result
		},
	})
}

// Reconcile is the entrypoint for incoming requests from k8s
func (r *SecretRequestReconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	log := r.log.WithValues("request", request)

	var secretRequest sgv1alpha1.SecretRequest

	err := r.client.Get(ctx, request.NamespacedName, &secretRequest)
	if err != nil {
		if errors.IsNotFound(err) {
			// Do not requeue as there is nothing to do when request is deleted
			return reconcile.Result{}, nil
		}
		// Requeue to try to fetch request again
		return reconcile.Result{Requeue: true}, err
	}

	if secretRequest.DeletionTimestamp != nil {
		// Do not requeue as there is nothing to do
		// Associated secret has owned ref so it's going to be deleted
		return reconcile.Result{}, nil
	}

	status := &reconciler.Status{
		secretRequest.Status.GenericStatus,
		func(st sgv1alpha1.GenericStatus) { secretRequest.Status.GenericStatus = st },
	}

	status.SetReconciling(secretRequest.ObjectMeta)

	reconcileResult, reconcileErr := status.WithReconcileCompleted(r.reconcile(ctx, secretRequest, log))

	err = r.updateStatus(ctx, secretRequest)
	if err != nil {
		return reconcile.Result{}, err
	}

	return reconcileResult, reconcileErr
}

func (r *SecretRequestReconciler) reconcile(
	ctx context.Context, secretRequest sgv1alpha1.SecretRequest,
	log logr.Logger) (reconcile.Result, error) {

	err := secretRequest.Validate()
	if err != nil {
		// Do not requeue as there is nothing this controller can do until secret request is fixed
		return reconcile.Result{}, reconciler.TerminalReconcileErr{err}
	}

	log.Info("Reconciling")

	matcher := SecretMatcher{
		FromName:      secretRequest.Name,
		FromNamespace: secretRequest.Spec.FromNamespace,
		ToNamespace:   secretRequest.Namespace,
	}

	secrets := r.secretExports.MatchedSecretsForImport(matcher)

	switch len(secrets) {
	case 0:
		err := r.deleteAssociatedSecret(ctx, secretRequest)
		if err != nil {
			// Requeue to try to delete a bit later
			return reconcile.Result{Requeue: true}, err
		}
		// Do not requeue since export is not offered
		return reconcile.Result{}, reconciler.TerminalReconcileErr{fmt.Errorf("No matching export/secret")}

	case 1:
		return r.copyAssociatedSecret(ctx, secretRequest, secrets[0])

	default:
		panic("Internal inconsistency: multiple exports/secrets matched one ns+name")
	}
}

func (r *SecretRequestReconciler) copyAssociatedSecret(
	ctx context.Context, req sgv1alpha1.SecretRequest, srcSecret *corev1.Secret) (reconcile.Result, error) {

	secret := reconciler.NewSecret(&req, nil)
	secret.ApplySecret(*srcSecret)

	err := r.client.Create(ctx, secret.AsSecret())
	switch {
	case err == nil:
		// Do not requeue since we copied secret successfully
		return reconcile.Result{}, nil

	case errors.IsAlreadyExists(err):
		var existingSecret corev1.Secret
		existingSecretNN := types.NamespacedName{Namespace: req.Namespace, Name: req.Name}

		err := r.client.Get(ctx, existingSecretNN, &existingSecret)
		if err != nil {
			// Requeue to try to fetch a bit later
			return reconcile.Result{Requeue: true}, fmt.Errorf("Getting imported secret: %s", err)
		}

		secret.AssociateExistingSecret(existingSecret)

		err = r.client.Update(ctx, secret.AsSecret())
		if err != nil {
			// Requeue to try to update a bit later
			return reconcile.Result{Requeue: true}, fmt.Errorf("Updating imported secret: %s", err)
		}

		// Do not requeue since we copied secret successfully
		return reconcile.Result{}, nil

	default:
		// Requeue to try to create a bit later
		return reconcile.Result{Requeue: true}, fmt.Errorf("Creating imported secret: %s", err)
	}
}

func (r *SecretRequestReconciler) deleteAssociatedSecret(
	ctx context.Context, req sgv1alpha1.SecretRequest) error {

	var secret corev1.Secret
	secretNN := types.NamespacedName{Namespace: req.Namespace, Name: req.Name}

	err := r.client.Get(ctx, secretNN, &secret)
	if err != nil {
		return nil
	}

	err = r.client.Delete(ctx, &secret)
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("Deleting associated secret: %s", err)
	}
	return nil
}

func (r *SecretRequestReconciler) updateStatus(
	ctx context.Context, req sgv1alpha1.SecretRequest) error {

	err := r.client.Status().Update(ctx, &req)
	if err != nil {
		return fmt.Errorf("Updating secret request status: %s", err)
	}
	return nil
}
