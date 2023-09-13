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

package openstackclient

import (
	env "github.com/openstack-k8s-operators/lib-common/modules/common/env"
	clientv1 "github.com/openstack-k8s-operators/openstack-operator/apis/client/v1beta1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

const (
	// ServiceCommand -
	ServiceCommand = "sudo -E /usr/local/bin/kolla_set_configs && sudo -E /usr/local/bin/kolla_start"
)

// ClientPod func
func ClientPod(
	instance *clientv1.OpenStackClient,
	labels map[string]string,
	configHash string,
) *corev1.Pod {

	clientPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name,
			Namespace: instance.Namespace,
		},
	}

	envVars := map[string]env.Setter{}
	envVars["KOLLA_CONFIG_STRATEGY"] = env.SetValue("COPY_ALWAYS")
	envVars["OS_CLOUD"] = env.SetValue("default")
	envVars["CONFIG_HASH"] = env.SetValue(configHash)

	clientPod.ObjectMeta = metav1.ObjectMeta{
		Name:      instance.Name,
		Namespace: instance.Namespace,
		Labels:    labels,
	}
	clientPod.Spec.TerminationGracePeriodSeconds = ptr.To[int64](0)
	clientPod.Spec.ServiceAccountName = instance.RbacResourceName()
	clientContainer := corev1.Container{
		Name:  "openstackclient",
		Image: instance.Spec.ContainerImage,
		Command: []string{
			"/bin/bash",
		},
		Args: []string{"-c", ServiceCommand},
		SecurityContext: &corev1.SecurityContext{
			RunAsUser:  ptr.To[int64](42401),
			RunAsGroup: ptr.To[int64](42401),
		},
		Env: env.MergeEnvs([]corev1.EnvVar{}, envVars),
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      "openstack-config",
				MountPath: "/var/lib/config-data/clouds.yaml",
				SubPath:   "clouds.yaml",
			},
			{
				Name:      "openstack-config-secret",
				MountPath: "/var/lib/config-data/secure.yaml",
				SubPath:   "secure.yaml",
			},
			{
				Name:      "config-data",
				MountPath: "/var/lib/kolla/config_files/config.json",
				SubPath:   "config.json",
				ReadOnly:  true,
			},
		},
	}
	if instance.Spec.CaSecretName != "" {
		clientContainer.VolumeMounts = append(clientContainer.VolumeMounts,
			corev1.VolumeMount{
				Name:      "ca",
				MountPath: "/var/lib/config-data/ca-certificates",
			})
	}

	clientPod.Spec.Containers = []corev1.Container{clientContainer}

	clientPod.Spec.Volumes = clientPodVolumes(instance, labels)
	if instance.Spec.NodeSelector != nil && len(instance.Spec.NodeSelector) > 0 {
		clientPod.Spec.NodeSelector = instance.Spec.NodeSelector
	}

	return clientPod
}

func clientPodVolumes(
	instance *clientv1.OpenStackClient,
	labels map[string]string,
) []corev1.Volume {
	volumes := []corev1.Volume{
		{
			Name: "openstack-config",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: instance.Spec.OpenStackConfigMap,
					},
				},
			},
		},
		{
			Name: "openstack-config-secret",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: instance.Spec.OpenStackConfigSecret,
				},
			},
		},
		{
			Name: "config-data",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: instance.Name + "-config-data",
					},
				},
			},
		},
	}

	if instance.Spec.CaSecretName != "" {
		volumes = append(volumes,
			corev1.Volume{
				Name: "ca",
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{
						SecretName: instance.Spec.CaSecretName,
					},
				},
			})
	}

	return volumes
}
