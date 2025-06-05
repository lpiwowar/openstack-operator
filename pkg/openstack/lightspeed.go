package openstack

import (
	"context"
	"path/filepath"

	"github.com/openstack-k8s-operators/lib-common/modules/common/helper"
	"github.com/openstack-k8s-operators/lib-common/modules/common/util"
	corev1beta1 "github.com/openstack-k8s-operators/openstack-operator/apis/core/v1beta1"
	"github.com/openstack-k8s-operators/openstack-operator/pkg/operator/bindata"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	OpenShiftLightspeedDeployment string = "lightspeed.yaml"
	OLSConfigFile                 string = "olsconfig.yaml"
)

func renderAndApply(
	file string,
	data *bindata.RenderData,
	instance *corev1beta1.OpenStackControlPlane,
	ctx *context.Context,
	helper *helper.Helper,
	setControllerReference bool,
) (ctrl.Result, error) {
	bindir := util.GetEnvVar("BASE_BINDATA", "/home/stack/lpiwowar/openstack-operator/bindata")
	//bindir := util.GetEnvVar("BASE_BINDATA", "/bindata")
	lightspeedBindir := filepath.Join(bindir, "lightspeed")

	lightspeedOperatorTemplate := filepath.Join(lightspeedBindir, file)
	objs, err := bindata.RenderTemplate(lightspeedOperatorTemplate, data)
	if err != nil {
		log.Log.Error(err, "failed applying manifests")
		return ctrl.Result{}, err
	}

	// Now apply the object
	for _, obj := range objs {
		if setControllerReference {
			log.Log.Info("Setting controller reference", "object", obj.GetName(), "controller", instance.Name)
			if err := controllerutil.SetControllerReference(instance, obj, helper.GetScheme()); err != nil {
				log.Log.Error(err, "failed setting controller reference")
				return ctrl.Result{}, errors.Wrapf(err, "failed setting controller reference")
			}
		} else {
			log.Log.Info("Skipping setting controller reference", "object", obj.GetName(), "controller", instance.Name)
		}

		err = bindata.ApplyObject(*ctx, helper.GetClient(), obj)
		if err != nil {
			return ctrl.Result{}, errors.Wrapf(err, "failed to apply object %v", obj)
		}
	}

	return ctrl.Result{}, nil
}

func ReconcileLightSpeed(
	ctx context.Context,
	instance *corev1beta1.OpenStackControlPlane,
	version *corev1beta1.OpenStackVersion,
	helper *helper.Helper,
) (ctrl.Result, error) {
	log.Log.Info("Running ReconcileLightSpeed!")

	if !instance.Spec.LightSpeed.Enabled {
		return reconcileLightSpeedDelete(ctx, instance, version, helper)
	}

	data := bindata.MakeRenderData()
	if _, err := renderAndApply(OpenShiftLightspeedDeployment, &data, instance, &ctx, helper, true); err != nil {
		return ctrl.Result{}, err
	}

	data = bindata.MakeRenderData()
	data.Data["LLMCredentials"] = instance.Spec.LightSpeed.LLMCredentials
	data.Data["LLMEndpoint"] = instance.Spec.LightSpeed.LLMEndpoint
	data.Data["TLSCert"] = instance.Spec.LightSpeed.TLSCert
	if _, err := renderAndApply(OLSConfigFile, &data, instance, &ctx, helper, false); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func reconcileLightSpeedDelete(
	ctx context.Context,
	instance *corev1beta1.OpenStackControlPlane,
	version *corev1beta1.OpenStackVersion,
	helper *helper.Helper,
) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}
