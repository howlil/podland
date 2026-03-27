// Package k8s provides Kubernetes management capabilities for VMs
package k8s

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// VMManager manages Kubernetes resources for VMs
type VMManager struct {
	clientset *kubernetes.Clientset
}

// VM represents a virtual machine configuration
type VM struct {
	ID        string
	UserID    string
	Name      string
	OS        string
	Tier      string
	CPU       float64
	RAM       int64
	Storage   int64
	Status    string
	Image     string
	Domain    string
	PublicKey string
}

// NewVMManager creates a new VMManager instance
func NewVMManager(kubeconfig string) (*VMManager, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %w", err)
	}

	return &VMManager{clientset: clientset}, nil
}

// NewVMManagerFromClusterConfig creates a VMManager using in-cluster config
func NewVMManagerFromClusterConfig() (*VMManager, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load in-cluster config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %w", err)
	}

	return &VMManager{clientset: clientset}, nil
}

// CreateVM creates all Kubernetes resources for a VM
func (m *VMManager) CreateVM(ctx context.Context, vm *VM) error {
	namespace := fmt.Sprintf("user-%s", vm.UserID)

	// 1. Ensure namespace exists
	err := m.ensureNamespace(ctx, namespace)
	if err != nil {
		return fmt.Errorf("failed to ensure namespace: %w", err)
	}

	// 2. Create PVC
	err = m.createPVC(ctx, namespace, vm)
	if err != nil {
		return fmt.Errorf("failed to create PVC: %w", err)
	}

	// 3. Create Deployment
	err = m.createDeployment(ctx, namespace, vm)
	if err != nil {
		return fmt.Errorf("failed to create Deployment: %w", err)
	}

	// 4. Create Service
	err = m.createService(ctx, namespace, vm)
	if err != nil {
		return fmt.Errorf("failed to create Service: %w", err)
	}

	// 5. Create Ingress (for HTTP/HTTPS)
	err = m.createIngress(ctx, namespace, vm)
	if err != nil {
		return fmt.Errorf("failed to create Ingress: %w", err)
	}

	return nil
}

// DeleteVM deletes all Kubernetes resources for a VM
func (m *VMManager) DeleteVM(ctx context.Context, vm *VM) error {
	namespace := fmt.Sprintf("user-%s", vm.UserID)
	name := fmt.Sprintf("vm-%s", vm.ID)

	// Delete in reverse order of creation

	// 1. Delete Ingress
	err := m.clientset.NetworkingV1().Ingresses(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to delete Ingress: %w", err)
	}

	// 2. Delete Service
	err = m.clientset.CoreV1().Services(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to delete Service: %w", err)
	}

	// 3. Delete Deployment
	err = m.clientset.AppsV1().Deployments(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to delete Deployment: %w", err)
	}

	// 4. Delete PVC
	err = m.clientset.CoreV1().PersistentVolumeClaims(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to delete PVC: %w", err)
	}

	return nil
}

// GetVMStatus returns the current status of a VM
func (m *VMManager) GetVMStatus(ctx context.Context, vm *VM) (string, error) {
	namespace := fmt.Sprintf("user-%s", vm.UserID)
	name := fmt.Sprintf("vm-%s", vm.ID)

	// Get Deployment to check replica status
	deployment, err := m.clientset.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return "not_found", nil
		}
		return "", fmt.Errorf("failed to get Deployment: %w", err)
	}

	// Check deployment status
	if deployment.Status.ReadyReplicas > 0 {
		return "running", nil
	}

	if deployment.Status.Replicas == 0 {
		return "stopped", nil
	}

	// Check if there are any failed pods
	if deployment.Status.UnavailableReplicas > 0 {
		return "error", nil
	}

	return "pending", nil
}

// StartVM starts a VM by scaling the Deployment to 1 replica
func (m *VMManager) StartVM(ctx context.Context, vm *VM) error {
	namespace := fmt.Sprintf("user-%s", vm.UserID)
	name := fmt.Sprintf("vm-%s", vm.ID)

	replicas := int32(1)

	_, err := m.clientset.AppsV1().Deployments(namespace).UpdateScale(
		ctx,
		name,
		&autoscalingv1.Scale{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			Spec: autoscalingv1.ScaleSpec{
				Replicas: replicas,
			},
		},
		metav1.UpdateOptions{},
	)

	if err != nil {
		return fmt.Errorf("failed to scale deployment: %w", err)
	}

	return nil
}

// StopVM stops a VM by scaling the Deployment to 0 replicas
func (m *VMManager) StopVM(ctx context.Context, vm *VM) error {
	namespace := fmt.Sprintf("user-%s", vm.UserID)
	name := fmt.Sprintf("vm-%s", vm.ID)

	replicas := int32(0)

	_, err := m.clientset.AppsV1().Deployments(namespace).UpdateScale(
		ctx,
		name,
		&autoscalingv1.Scale{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			Spec: autoscalingv1.ScaleSpec{
				Replicas: replicas,
			},
		},
		metav1.UpdateOptions{},
	)

	if err != nil {
		return fmt.Errorf("failed to scale deployment: %w", err)
	}

	return nil
}

// ensureNamespace creates a namespace if it doesn't exist
func (m *VMManager) ensureNamespace(ctx context.Context, name string) error {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"pod-security.kubernetes.io/enforce": "restricted",
				"pod-security.kubernetes.io/audit":   "restricted",
				"pod-security.kubernetes.io/warn":    "restricted",
			},
		},
	}

	_, err := m.clientset.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
	if errors.IsAlreadyExists(err) {
		return nil
	}

	return err
}

// createPVC creates a PersistentVolumeClaim for VM storage
func (m *VMManager) createPVC(ctx context.Context, namespace string, vm *VM) error {
	storageClassName := "local-lvm"

	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("vm-%s-pvc", vm.ID),
			Namespace: namespace,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			StorageClassName: &storageClassName,
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse(fmt.Sprintf("%d", vm.Storage)),
				},
			},
		},
	}

	_, err := m.clientset.CoreV1().PersistentVolumeClaims(namespace).Create(ctx, pvc, metav1.CreateOptions{})
	if errors.IsAlreadyExists(err) {
		return nil
	}

	return err
}

// createDeployment creates a Deployment for the VM
func (m *VMManager) createDeployment(ctx context.Context, namespace string, vm *VM) error {
	trueVal := true
	falseVal := false

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("vm-%s", vm.ID),
			Namespace: namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": fmt.Sprintf("vm-%s", vm.ID),
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": fmt.Sprintf("vm-%s", vm.ID),
					},
				},
				Spec: corev1.PodSpec{
					SecurityContext: &corev1.PodSecurityContext{
						FSGroup: int64Ptr(1000),
					},
					Containers: []corev1.Container{
						{
							Name:  "vm",
							Image: vm.Image,
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse(fmt.Sprintf("%.2f", vm.CPU)),
									corev1.ResourceMemory: resource.MustParse(fmt.Sprintf("%d", vm.RAM)),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse(fmt.Sprintf("%.2f", vm.CPU)),
									corev1.ResourceMemory: resource.MustParse(fmt.Sprintf("%d", vm.RAM)),
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "vm-storage",
									MountPath: "/data",
								},
							},
							SecurityContext: &corev1.SecurityContext{
								RunAsNonRoot:             &trueVal,
								RunAsUser:                int64Ptr(1000),
								RunAsGroup:               int64Ptr(1000),
								AllowPrivilegeEscalation: &falseVal,
								Capabilities: &corev1.Capabilities{
									Drop: []corev1.Capability{"ALL"},
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "vm-storage",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: fmt.Sprintf("vm-%s-pvc", vm.ID),
								},
							},
						},
					},
				},
			},
		},
	}

	_, err := m.clientset.AppsV1().Deployments(namespace).Create(ctx, deployment, metav1.CreateOptions{})
	if errors.IsAlreadyExists(err) {
		return nil
	}

	return err
}

// createService creates a Service for SSH access to the VM
func (m *VMManager) createService(ctx context.Context, namespace string, vm *VM) error {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("vm-%s-ssh", vm.ID),
			Namespace: namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": fmt.Sprintf("vm-%s", vm.ID),
			},
			Ports: []corev1.ServicePort{
				{
					Name:       "ssh",
					Port:       22,
					TargetPort: intstr.FromInt(22),
					Protocol:   corev1.ProtocolTCP,
				},
			},
			Type: corev1.ServiceTypeClusterIP,
		},
	}

	_, err := m.clientset.CoreV1().Services(namespace).Create(ctx, service, metav1.CreateOptions{})
	if errors.IsAlreadyExists(err) {
		return nil
	}

	return err
}

// createIngress creates an Ingress for HTTP/HTTPS access to the VM
func (m *VMManager) createIngress(ctx context.Context, namespace string, vm *VM) error {
	ingressClassName := "traefik"

	ingress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("vm-%s", vm.ID),
			Namespace: namespace,
			Annotations: map[string]string{
				"traefik.ingress.kubernetes.io/router.entrypoints": "web,websecure",
				"traefik.ingress.kubernetes.io/router.tls":         "true",
			},
		},
		Spec: networkingv1.IngressSpec{
			IngressClassName: &ingressClassName,
			Rules: []networkingv1.IngressRule{
				{
					Host: fmt.Sprintf("%s.%s", vm.Name, vm.Domain),
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     "/",
									PathType: pathTypePtr(networkingv1.PathTypePrefix),
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: fmt.Sprintf("vm-%s-ssh", vm.ID),
											Port: networkingv1.ServiceBackendPort{
												Number: 80,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	_, err := m.clientset.NetworkingV1().Ingresses(namespace).Create(ctx, ingress, metav1.CreateOptions{})
	if errors.IsAlreadyExists(err) {
		return nil
	}

	return err
}

// Helper functions
func int32Ptr(i int32) *int32 { return &i }
func int64Ptr(i int64) *int64 { return &i }
func boolPtr(b bool) *bool    { return &b }
func pathTypePtr(t networkingv1.PathType) *networkingv1.PathType { return &t }
