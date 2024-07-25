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

package controller

import (
	"context"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	v1 "example.com/jsonserver/api/v1"
)

const (
	JsonServerFinalizerName         = "example.com/finalizer"
	JsonConfigKey                   = "db.json"
	JsonConfigMapMountPath          = "/data"
	JsonConfigPath                  = JsonConfigMapMountPath + "/" + JsonConfigKey
	JsonServerMatchLabelsKey        = "app"
	JsonServerImage                 = "backplane/json-server"
	JsonServerContainerName         = "json-server"
	JsonServerVolumeName            = "json-config"
	JsonServerContainerPort         = 3000
	JsonServerContainerPortName     = "http"
	JsonServerContainerPortProtocol = "TCP"
)

// JsonServerReconciler reconciles a JsonServer object
type JsonServerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=example.com,resources=jsonservers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=example.com,resources=jsonservers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=example.com,resources=jsonservers/finalizers,verbs=update
// +kubebuilder:rbac:groups=resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *JsonServerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	jsonServer := &v1.JsonServer{}

	if err := r.Get(ctx, req.NamespacedName, jsonServer); err != nil {
		if apierrors.IsNotFound(err) {
			log.Log.Info("JsonServer not found", "NamespacedName", req.NamespacedName)
			return ctrl.Result{}, nil
		}

		log.Log.Error(err, "Failed to fetch JsonServer",
			"NamespacedName", req.NamespacedName)
		return ctrl.Result{}, err
	}

	configMap := &corev1.ConfigMap{}

	if err := r.Get(ctx, req.NamespacedName, configMap); err != nil {
		if apierrors.IsNotFound(err) == false {
			log.Log.Error(err, "Failed to fetch ConfigMap",
				"NamespacedName", req.NamespacedName)
			return ctrl.Result{}, err
		}

		log.Log.Info("ConfigMap not found. Creating...", "NamespacedName", req.NamespacedName)

		if err := r.createConfigMap(ctx, req, jsonServer); err != nil {
			if apierrors.IsAlreadyExists(err) == false {
				log.Log.Error(err, "Failed to create ConfigMap", "NamespacedName", req.NamespacedName)
				return ctrl.Result{}, err
			}

			log.Log.Info("ConfigMap already exist", "NamespacedName", req.NamespacedName)
			//TODO check owner reference, if JsonServer isn't owner - return error
		} else {
			log.Log.Info("Successfully created ConfigMap", "NamespacedName", req.NamespacedName)
			return ctrl.Result{}, nil
		}
	}

	deployment := &appsv1.Deployment{}

	if err := r.Get(ctx, req.NamespacedName, deployment); err != nil {
		if apierrors.IsNotFound(err) == false {
			log.Log.Error(err, "Failed to fetch Deployment",
				"NamespacedName", req.NamespacedName)
			return ctrl.Result{}, err
		}

		log.Log.Info("Deployment not found. Creating...", "NamespacedName", req.NamespacedName)

		if err := r.createDeployment(ctx, req, jsonServer); err != nil {
			if apierrors.IsAlreadyExists(err) == false {
				log.Log.Error(err, "Failed to create Deployment", "NamespacedName", req.NamespacedName)
				return ctrl.Result{}, err
			}

			log.Log.Info("ConfigMap already exist", "NamespacedName", req.NamespacedName)
			//TODO check owner reference, if JsonServer isn't owner - return error
		} else {
			log.Log.Info("Successfully created Deployment", "NamespacedName", req.NamespacedName)
			return ctrl.Result{}, nil
		}
	}

	service := &corev1.Service{}

	if err := r.Get(ctx, req.NamespacedName, service); err != nil {
		if apierrors.IsNotFound(err) == false {
			log.Log.Error(err, "Failed to fetch Service", "NamespacedName", req.NamespacedName)
			return ctrl.Result{}, err
		}

		log.Log.Info("Service not found. Creating...", "NamespacedName", req.NamespacedName)

		if err := r.createService(ctx, req, jsonServer); err != nil {
			if apierrors.IsAlreadyExists(err) == false {
				log.Log.Error(err, "Failed to create Service", "NamespacedName", req.NamespacedName)
				return ctrl.Result{}, err
			}

			log.Log.Info("Service already exist", "NamespacedName", req.NamespacedName)
			//TODO check owner reference, if JsonServer isn't owner - return error
		} else {
			log.Log.Info("Successfully created Service", "NamespacedName", req.NamespacedName)
			return ctrl.Result{}, nil
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *JsonServerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.JsonServer{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&corev1.Service{}).
		Complete(r)
}

func (r *JsonServerReconciler) createConfigMap(ctx context.Context, req ctrl.Request, jsonServer *v1.JsonServer) error {
	configMap := r.generateConfigMap(jsonServer)

	if err := ctrl.SetControllerReference(jsonServer, configMap, r.Scheme); err != nil {
		log.Log.Error(err, "Failed to set owner reference on Configmap", "NamespacedName", req.NamespacedName)

		return err
	}

	return r.Create(ctx, configMap)
}

func (r *JsonServerReconciler) generateConfigMap(jsonServer *v1.JsonServer) *corev1.ConfigMap {
	//TODO I'd copy JsonServer labels also
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jsonServer.Name,
			Namespace: jsonServer.Namespace,
		},
		Data: map[string]string{JsonConfigKey: jsonServer.Spec.JsonConfig},
	}
}

func (r *JsonServerReconciler) createDeployment(ctx context.Context, req ctrl.Request, jsonServer *v1.JsonServer) error {
	deployment := r.generateDeployment(jsonServer)

	if err := ctrl.SetControllerReference(jsonServer, deployment, r.Scheme); err != nil {
		log.Log.Error(err, "Failed to set owner reference on Deployment", "NamespacedName", req.NamespacedName)

		return err
	}

	return r.Create(ctx, deployment)
}

func (r *JsonServerReconciler) generateDeployment(jsonServer *v1.JsonServer) *appsv1.Deployment {
	//TODO I'd copy JsonServer labels also
	matchLabels := client.MatchingLabels{
		JsonServerMatchLabelsKey: jsonServer.Name,
	}

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jsonServer.Name,
			Namespace: jsonServer.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: jsonServer.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: matchLabels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: matchLabels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  JsonServerContainerName,
							Image: JsonServerImage,
							Args:  []string{JsonConfigPath},
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: JsonServerContainerPort,
									Name:          JsonServerContainerPortName,
									Protocol:      JsonServerContainerPortProtocol,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      JsonServerVolumeName,
									MountPath: JsonConfigMapMountPath,
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: JsonServerVolumeName,
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: jsonServer.Name,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (r *JsonServerReconciler) createService(ctx context.Context, req ctrl.Request, jsonServer *v1.JsonServer) error {
	configMap := r.generateService(jsonServer)

	if err := ctrl.SetControllerReference(jsonServer, configMap, r.Scheme); err != nil {
		log.Log.Error(err, "Failed to set owner reference on Configmap", "NamespacedName", req.NamespacedName)

		return err
	}

	return r.Create(ctx, configMap)
}

func (r *JsonServerReconciler) generateService(jsonServer *v1.JsonServer) *corev1.Service {
	matchLabels := client.MatchingLabels{
		JsonServerMatchLabelsKey: jsonServer.Name,
	}
	//TODO I'd copy JsonServer labels also

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jsonServer.Name,
			Namespace: jsonServer.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: matchLabels,
			Ports: []corev1.ServicePort{
				{
					Protocol:   JsonServerContainerPortProtocol,
					Port:       JsonServerContainerPort,
					TargetPort: intstr.IntOrString{IntVal: JsonServerContainerPort},
				},
			},
		},
	}
}
