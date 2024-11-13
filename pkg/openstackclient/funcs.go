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
	"context"
	"fmt"
	"slices"

	env "github.com/openstack-k8s-operators/lib-common/modules/common/env"
	"github.com/openstack-k8s-operators/lib-common/modules/common/helper"
	"github.com/openstack-k8s-operators/lib-common/modules/common/tls"
	clientv1 "github.com/openstack-k8s-operators/openstack-operator/apis/client/v1beta1"
	telemetryv1 "github.com/openstack-k8s-operators/telemetry-operator/api/v1beta1"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ClientPodSpec func
func SetClientPodSpec(
	ctx context.Context,
	instance *clientv1.OpenStackClient,
	helper *helper.Helper,
	configHash string,
	podSpec *corev1.PodSpec,
) {
	envVars := map[string]env.Setter{}
	envVars["OS_CLOUD"] = env.SetValue("default")
	envVars["CONFIG_HASH"] = env.SetValue(configHash)
	envVars["PROMETHEUS_HOST"] = env.SetValue(fmt.Sprintf("%s-prometheus.%s.svc",
		telemetryv1.DefaultServiceName,
		instance.Namespace))
	envVars["PROMETHEUS_PORT"] = env.SetValue(fmt.Sprint(telemetryv1.DefaultPrometheusPort))
	metricStorage := &telemetryv1.MetricStorage{}
	err := helper.GetClient().Get(ctx, client.ObjectKey{
		Namespace: instance.Namespace,
		Name:      telemetryv1.DefaultServiceName,
	}, metricStorage)
	if err == nil && metricStorage.Spec.PrometheusTLS.Enabled() {
		envVars["PROMETHEUS_CA_CERT"] = env.SetValue(tls.DownstreamTLSCABundlePath)
	}

	// create Volume and VolumeMounts
	volumes := clientPodVolumes(instance)
	volumeMounts := clientPodVolumeMounts()

	// add CA cert if defined
	if instance.Spec.CaBundleSecretName != "" {
		volumes = append(volumes, instance.Spec.CreateVolume())
		volumeMounts = append(volumeMounts, instance.Spec.CreateVolumeMounts(nil)...)
	}

	podSpec.TerminationGracePeriodSeconds = ptr.To[int64](0)
	podSpec.ServiceAccountName = instance.RbacResourceName()
	for _, volume := range volumes {
		idx := slices.IndexFunc(podSpec.Volumes, func(v corev1.Volume) bool {
			return v.Name == volume.Name
		})
		if idx == -1 {
			podSpec.Volumes = append(podSpec.Volumes, volume)
		} else {
			podSpec.Volumes[idx] = volume
		}
	}

	if len(podSpec.Containers) < 1 {
		podSpec.Containers = []corev1.Container{
			{
				Name: "openstackclient",
			},
		}
	}
	podSpec.Containers[0].Name = "openstackclient"
	podSpec.Containers[0].Image = instance.Spec.ContainerImage
	podSpec.Containers[0].Command = []string{"/bin/sleep"}
	podSpec.Containers[0].Args = []string{"infinity"}
	if podSpec.Containers[0].SecurityContext == nil {
		podSpec.Containers[0].SecurityContext = &corev1.SecurityContext{}
	}
	podSpec.Containers[0].SecurityContext.RunAsUser = ptr.To[int64](42401)
	podSpec.Containers[0].SecurityContext.RunAsGroup = ptr.To[int64](42401)
	podSpec.Containers[0].SecurityContext.RunAsNonRoot = ptr.To(true)
	podSpec.Containers[0].SecurityContext.AllowPrivilegeEscalation = ptr.To(false)
	if podSpec.Containers[0].SecurityContext.Capabilities == nil {
		podSpec.Containers[0].SecurityContext.Capabilities = &corev1.Capabilities{
			Drop: []corev1.Capability{},
		}
	}
	{
		idx := slices.Index(podSpec.Containers[0].SecurityContext.Capabilities.Drop, "ALL")
		if idx == -1 {
			podSpec.Containers[0].SecurityContext.Capabilities.Drop = append(podSpec.Containers[0].SecurityContext.Capabilities.Drop, "ALL")
		}
	}
	podSpec.Containers[0].Env = env.MergeEnvs([]corev1.EnvVar{}, envVars)
	for _, volumeMount := range volumeMounts {
		idx := slices.Index(podSpec.Containers[0].VolumeMounts, volumeMount)
		if idx == -1 {
			podSpec.Containers[0].VolumeMounts = append(podSpec.Containers[0].VolumeMounts, volumeMount)
		} else {
			podSpec.Containers[0].VolumeMounts[idx] = volumeMount
		}
	}

	if instance.Spec.NodeSelector != nil {
		podSpec.NodeSelector = *instance.Spec.NodeSelector
	} else {
		podSpec.NodeSelector = nil
	}
}

func clientPodVolumeMounts() []corev1.VolumeMount {
	return []corev1.VolumeMount{
		{
			Name:      "openstack-config",
			MountPath: "/home/cloud-admin/.config/openstack/clouds.yaml",
			SubPath:   "clouds.yaml",
		},
		{
			Name:      "openstack-config-secret",
			MountPath: "/home/cloud-admin/.config/openstack/secure.yaml",
			SubPath:   "secure.yaml",
		},
		{
			Name:      "openstack-config-secret",
			MountPath: "/home/cloud-admin/cloudrc",
			SubPath:   "cloudrc",
		},
	}
}

func clientPodVolumes(
	instance *clientv1.OpenStackClient,
) []corev1.Volume {
	return []corev1.Volume{
		{
			Name: "openstack-config",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: *instance.Spec.OpenStackConfigMap,
					},
				},
			},
		},
		{
			Name: "openstack-config-secret",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: *instance.Spec.OpenStackConfigSecret,
				},
			},
		},
	}
}
