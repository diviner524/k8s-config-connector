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

package main

import (
	"flag"
	"log"
	"os"

	"gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	infrav1alpha1 "pubsub-controller/api/v1alpha1"
	"pubsub-controller/controllers"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)

	_ = infrav1alpha1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	yamlFile, err := os.ReadFile("file.yaml")

	if err != nil {
		log.Fatalf("Error reading YAML file: %s\n", err)
	}

	// Unmarshal into ResourcesWrapper
	var resourcesWrapper controllers.ResourcesWrapper
	err = yaml.Unmarshal(yamlFile, &resourcesWrapper)
	if err != nil {
		log.Fatalf("Error unmarshalling YAML: %s\n", err)
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		Port:               9443,
		LeaderElection:     enableLeaderElection,
		LeaderElectionID:   "81c54d15.sample.org",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&controllers.PubSubReconciler{
		Client:           mgr.GetClient(),
		Log:              ctrl.Log.WithName("controllers").WithName("PubSub"),
		Scheme:           mgr.GetScheme(),
		OrderedResources: sort(convertToMap(resourcesWrapper)),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "PubSub")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

// convertToMap() takes a list of Resources as ResourceWrapper and returnes a map of resource name as key and ManagedResources as value
func convertToMap(resourcesWrapper controllers.ResourcesWrapper) map[string]*controllers.ManagedResource {
	resourcesMap := make(map[string]*controllers.ManagedResource)
	for _, resource := range resourcesWrapper.Resources {
		resourceCopy := resource // Create a copy of 'resource'
		managedResource := controllers.ManagedResource{
			Resource: &resourceCopy,
		}
		resourcesMap[resource.Name] = &managedResource
	}
	// This second loop parse the DependsOn field of each resource and add it to the dependencies of current resource
	for _, resource := range resourcesWrapper.Resources {
		for _, dependency := range resource.DependsOn {
			resourcesMap[resource.Name].Dependencies = append(resourcesMap[resource.Name].Dependencies, resourcesMap[dependency])
		}
	}
	return resourcesMap
}

// sort() takes a map of MangedResources and does a topological sort on it, then returns a list of *ManagedResources based on the topological order
func sort(resourcesMap map[string]*controllers.ManagedResource) []*controllers.ManagedResource {
	var sortedResources []*controllers.ManagedResource
	visited := make(map[string]bool)

	var visit func(string)
	visit = func(name string) {
		if visited[name] {
			return
		}
		visited[name] = true

		resource := resourcesMap[name]
		for _, dep := range resource.Dependencies {
			visit(dep.Resource.Name)
		}

		sortedResources = append(sortedResources, resource)
	}

	for name := range resourcesMap {
		visit(name)
	}

	return sortedResources
}
