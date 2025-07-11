/*
Copyright © 2022 Christian Hernandez christian@chernand.io

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
package cmd

import (
	"context"
	"fmt"
	"os"
	"strconv"

	argov1alpha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	yaml "sigs.k8s.io/yaml"

	"github.com/akuity/mta/pkg/argo"
	"github.com/akuity/mta/pkg/utils"
	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/cli-runtime/pkg/printers"
	client "sigs.k8s.io/controller-runtime/pkg/client"
)

// helmreleaseCmd represents the helmrelease command
var helmreleaseCmd = &cobra.Command{
	Use:     "helmrelease",
	Aliases: []string{"HelmRelease", "hr", "helmreleases"},
	Short:   "Exports a HelmRelease into an Application",
	Long: `This migration tool helps you move your Flux HelmReleases into Argo CD
Applications. Example:

mta helmrelease --name=myhelmrelease --namespace=flux-system | kubectl apply -n argocd -f -

This utilty exports the named HelmRelease and the source Helm repo and
creates a manifests to stdout, which you can pipe into an apply command
with kubectl.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get the Argo CD namespace
		argoCDNamespace, err := cmd.Flags().GetString("argocd-namespace")
		if err != nil {
			log.Fatal(err)
		}

		// Get the options from the CLI
		kubeConfig, err := cmd.Flags().GetString("kubeconfig")
		if err != nil {
			log.Fatal(err)
		}
		helmReleaseName, _ := cmd.Flags().GetString("name")
		helmReleaseNamespace, _ := cmd.Flags().GetString("namespace")
		confirmMigrate, _ := cmd.Flags().GetBool("confirm-migrate")
		argoProject, _ := cmd.Flags().GetString("argoproject")

		ctx := context.TODO()

		// Set up the schema because HelmRelease and Repo is a CRD
		scheme := runtime.NewScheme()
		_ = helmv2.AddToScheme(scheme)
		_ = sourcev1.AddToScheme(scheme)
		_ = argov1alpha1.AddToScheme(scheme)

		// create rest config using the kubeconfig file.
		restConfig, err := utils.NewRestConfig(kubeConfig)
		if err != nil {
			log.Fatal(err)
		}
		k, err := client.New(restConfig, client.Options{Scheme: scheme})
		if err != nil {
			log.Fatal(err)
		}

		//Get the helmrelease based on type, report if there's an error
		helmRelease := &helmv2.HelmRelease{}
		err = k.Get(ctx, types.NamespacedName{Namespace: helmReleaseNamespace, Name: helmReleaseName}, helmRelease)
		if err != nil {
			log.Fatal(err)
		}

		var sourceRefName string
		var helmChartNameSafe string
		var helmTargetRevision string

		sourceRefName = helmRelease.Spec.Chart.Spec.SourceRef.Name
		helmChartNameSafe = helmRelease.Spec.Chart.Spec.Chart
		helmTargetRevision = helmRelease.Spec.Chart.Spec.Version

		helmRepoNamespace := helmRelease.Spec.Chart.Spec.SourceRef.Namespace
		if helmRepoNamespace == "" {
			helmRepoNamespace = helmRelease.Namespace
		}

		helmRepo := &sourcev1.HelmRepository{}
		err = k.Get(ctx, types.NamespacedName{Namespace: helmRepoNamespace, Name: sourceRefName}, helmRepo)
		if err != nil {
			log.Fatal(err)
		}

		// Get the helmchart based on type, report if error
		helmChartName := fmt.Sprintf("%s-%s", helmReleaseNamespace, helmReleaseName)

		helmChart := &sourcev1.HelmChart{}
		err = k.Get(ctx, types.NamespacedName{Namespace: helmRepoNamespace, Name: helmChartName}, helmChart)
		if err != nil {
			log.Fatal(err)
		}

		var valuesYaml []byte
		if helmRelease.Spec.Values != nil {
			valuesYaml, _ = yaml.Marshal(helmRelease.Spec.Values)
		} else {
			valuesYaml = []byte{}
		}

		createNamespace := "false"
		if helmRelease.Spec.Install != nil {
			createNamespace = strconv.FormatBool(helmRelease.Spec.Install.CreateNamespace)
		}

		helmAppNamePrefix := helmRelease.Spec.TargetNamespace
		if helmAppNamePrefix == "" {
			helmAppNamePrefix = helmReleaseNamespace
		}

		helmRepoURL := helmRepo.Spec.URL

		helmApp := argo.ArgoCdHelmApplication{
			Name:                 helmRelease.Name + helmAppNamePrefix,
			Namespace:            argoCDNamespace,
			DestinationNamespace: helmRelease.Spec.TargetNamespace,
			DestinationServer:    "https://kubernetes.default.svc",
			Project:              argoProject,
			HelmChart:            helmChartNameSafe,
			HelmRepo:             helmRepoURL,
			HelmTargetRevision:   helmTargetRevision,
			HelmValues:           string(valuesYaml),
			HelmCreateNamespace:  createNamespace,
		}

		helmArgoCdApp, err := argo.GenArgoCdHelmApplication(helmApp)
		if err != nil {
			log.Fatal(err)
		}

		if confirmMigrate {
			log.Infof("Migrating HelmRelease %q to Argo CD Application", helmRelease.Name)
			_ = utils.SuspendFluxObject(k, ctx, helmRelease)
			_ = utils.SuspendFluxObject(k, ctx, helmRepo)
			_ = utils.SuspendFluxObject(k, ctx, helmChart)
			_ = utils.CreateK8SObjects(k, ctx, helmArgoCdApp)
			_ = utils.DeleteK8SObjects(k, ctx, helmRelease)
			_ = utils.DeleteK8SObjects(k, ctx, helmRepo)
			_ = utils.DeleteK8SObjects(k, ctx, helmChart)
		} else {
			printr := printers.NewTypeSetter(k.Scheme()).ToPrinter(&printers.YAMLPrinter{})
			_ = printr.PrintObj(helmArgoCdApp, os.Stdout)
		}
	},
}

func GetHelmRepoNamespace(helmRelease *helmv2.HelmRelease) string {
	helmRepoNamespace := helmRelease.Spec.Chart.Spec.SourceRef.Namespace
	if helmRepoNamespace == "" {
		helmRepoNamespace = helmRelease.Namespace
	}

	return helmRepoNamespace
}

func init() {
	rootCmd.AddCommand(helmreleaseCmd)
	rootCmd.MarkPersistentFlagRequired("name")

	helmreleaseCmd.Flags().Bool("confirm-migrate", false, "Automatically Migrate the HelmRelease to an ApplicationSet")
}
