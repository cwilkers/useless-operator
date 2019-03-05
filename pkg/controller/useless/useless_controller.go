package useless

import (
	"context"
        "reflect"
        "time"

	uselessv1alpha1 "github.com/cwilkers/useless-operator/pkg/apis/useless/v1alpha1"

        appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
        "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_useless")

// Add creates a new Useless Controller and adds it to the Manager. The Manager
// will set fields on the Controller and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileUseless{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("useless-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Useless
	err = c.Watch(&source.Kind{Type: &uselessv1alpha1.Useless{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource Pods and requeue the owner Useless
	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &uselessv1alpha1.Useless{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileUseless{}

// ReconcileUseless reconciles a Useless object
type ReconcileUseless struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Useless object and makes changes based on the state read
// and what is in the Useless.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileUseless) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Useless")

	// Fetch the Useless instance
	instance := &uselessv1alpha1.Useless{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
                        reqLogger.Info("Useless resource not found. Ignoring")
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
                reqLogger.Error(err, "Failed to get Useless")
		return reconcile.Result{}, err
	}

	// Check if this deployment already exists
	found := &appsv1.Deployment{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
	        // Define a new Deployment
	        dep := r.deploymentForUseless(instance)
		reqLogger.Info("Creating a new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
		err = r.client.Create(context.TODO(), dep)
		if err != nil {
                        reqLogger.Error(err, "Failed to create new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
			return reconcile.Result{}, err
		}

		// Deployment created successfully - return
                return reconcile.Result{}, nil
	} else if err != nil {
                reqLogger.Error(err, "Failed to get Deployment")
		return reconcile.Result{}, err
	}

        // Ensure the deployment size is the same as the spec
        size := instance.Spec.Size
        if *found.Spec.Replicas != size {
		reqLogger.Info("Reconciling Deployment replicas to CR size")
                found.Spec.Replicas = &size
                err = r.client.Update(context.TODO(), found)
                if err != nil {
                        reqLogger.Error(err, "Failed to update Deployment", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)
                        return reconcile.Result{}, err
                }
                time.Sleep(time.Second*5)
                // Deployment updated, requeue to get list of pods in a minute
                return reconcile.Result{RequeueAfter: time.Minute}, nil
        }

        // Update the Useless CR status with pod names
        podList := &corev1.PodList{}
        labelSelector := labels.SelectorFromSet(labelsForUseless(instance.Name))
        listOps := &client.ListOptions{Namespace: instance.Namespace, LabelSelector: labelSelector}
        err = r.client.List(context.TODO(), listOps, podList)
        if err != nil {
                reqLogger.Error(err, "Failed to list pods", "Useless.Namespace", instance.Namespace, "Useless.Name", instance.Name)
                return reconcile.Result{}, err
        }
        podNames := getPodNames(podList.Items)

        // Update status.Nodes if needed
        if !reflect.DeepEqual(podNames, instance.Status.Nodes) {
		reqLogger.Info("Setting status.nodes to current pod list")
                instance.Status.Nodes = podNames
                err := r.client.Status().Update(context.TODO(), instance)
                if err != nil {
                        reqLogger.Error(err, "Failed to update Useless status")
                        return reconcile.Result{}, err
                }
                time.Sleep(time.Second*5)
                return reconcile.Result{}, nil
        }

        // Now that Deployment has been reconciled, reset the CR's size to zero if necessary to shut off
        // useless machine

        if size > 0 {
		reqLogger.Info("Zero-ing out CR size param")
                instance.Spec.Size = 0
                err := r.client.Update(context.TODO(), instance)
                if err != nil {
                        reqLogger.Error(err, "Failed to update Useless spec.size")
                        return reconcile.Result{}, err
                }
                time.Sleep(time.Second*5)
                return reconcile.Result{}, nil
        }

        return reconcile.Result{}, nil

}

// deploymentForUseless returns a deployment of multiple busybox pods with the same name/namespace as the cr
func (r *ReconcileUseless) deploymentForUseless(cr *uselessv1alpha1.Useless) *appsv1.Deployment {
	labelMap := labelsForUseless(cr.Name)
        replicas := cr.Spec.Size

        dep := &appsv1.Deployment{
                TypeMeta: metav1.TypeMeta{
                    APIVersion: "apps/v1",
                    Kind:       "Deployment",
                },
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
		},
                Spec: appsv1.DeploymentSpec{
                        Replicas: &replicas,
                        Selector: &metav1.LabelSelector{
                                MatchLabels: labelMap,
                        },
		        Template: corev1.PodTemplateSpec{
		                ObjectMeta: metav1.ObjectMeta{
			                Labels:    labelMap,
                                },
                                Spec: corev1.PodSpec{
			                Containers: []corev1.Container{{
					        Name:    "busybox",
					        Image:   "busybox",
					        Command: []string{"sleep", "3600"},
                                        }},
                                },
			},
		},
	}
        controllerutil.SetControllerReference(cr, dep, r.scheme)
        return dep
}

// labelsForUseless returns the labels for selecting pod resources given the name
func labelsForUseless(name string) map[string]string {
        return map[string]string{"app": "useless", "useless_cr": name}
}

// getPodNames returns the pod names of the array of pods passed in
func getPodNames(pods []corev1.Pod) []string {
        var podNames []string
        for _, pod := range pods {
                podNames = append(podNames, pod.Name)
        }
        return podNames
}
// vim: tabstop=8 shiftwidth=8
