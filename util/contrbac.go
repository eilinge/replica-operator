package util

import (
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func MakeRBACObjects(fbName, fbNamespace string) (rbacv1.ClusterRole, corev1.ServiceAccount, rbacv1.ClusterRoleBinding) {
	cr := rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: "replica-controller",
		},
		Rules: []rbacv1.PolicyRule{
			{
				Verbs:     []string{"get,update,create,delete"},
				APIGroups: []string{""},
				Resources: []string{"pods,deployments"},
			},
		},
	}

	sa := corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fbName,
			Namespace: fbNamespace,
		},
	}

	crb := rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: "replica-controller",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      rbacv1.ServiceAccountKind,
				Name:      fbName,
				Namespace: fbNamespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "ClusterRole",
			Name:     "replica-controller",
		},
	}

	return cr, sa, crb
}
