/*
Copyright 2023-2024 Bull SAS

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
package service

import (
	"errors"
	"etsn/server/jobmanager-service/models"
	"etsn/server/jobmanager-service/utils/logs"
	"fmt"

	"gopkg.in/yaml.v3"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

// ManifestService provides methods to process and manage manifests.
type ManifestService interface {
	ProcessManifests(comp models.ComponentDTO, manifests []map[string]interface{}) ([]models.Content, error)
}

type manifestService struct{}

// NewManifestService creates a new instance of ManifestService.
func NewManifestService() ManifestService {
	return &manifestService{}
}

// ProcessManifests processes manifests related to a component and returns a map of plain manifests.
func (ms *manifestService) ProcessManifests(comp models.ComponentDTO, manifests []map[string]interface{}) ([]models.Content, error) {
	contentList := make([]models.Content, 0)
	for _, manifest := range manifests {
		manifestMap, ok := manifest["metadata"].(map[interface{}]interface{})
		if !ok {
			return nil, errors.New("invalid manifest structure")
		}

		manifestName, ok := manifestMap["name"].(string)
		if !ok {
			return nil, errors.New("invalid manifest name")
		}

		for _, manifestHeader := range comp.Manifests {
			if manifestHeader.Name == manifestName {
				manifestYAML, err := yaml.Marshal(manifest)
				if err != nil {
					logs.Logger.Println("ERROR during k8s manifest marshalling: " + err.Error())
					continue
				}

				_, err = ms.validateK8sManifest(string(manifestYAML))
				if err != nil {
					logs.Logger.Println("ERROR during k8s manifest validation: " + err.Error())
					continue
				}

				logs.Logger.Println("Manifest content: " + string(manifestYAML))
				content := models.Content{
					Name: manifestHeader.Name,
					Yaml: string(manifestYAML),
				}
				contentList = append(contentList, content)
				// plainManifestMap[manifestHeader.Name] = models.PlainManifest{
				// 	YamlString: string(manifestYAML),
				// }
			}
		}
	}

	return contentList, nil
}

func (ms *manifestService) validateK8sManifest(yamlString string) (runtime.Object, error) {
	scheme := runtime.NewScheme()
	if err := appsv1.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("failed to add apps/v1 to scheme: %w", err)
	}
	if err := corev1.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("failed to add core/v1 to scheme: %w", err)
	}

	codecFactory := serializer.NewCodecFactory(scheme)
	decoder := codecFactory.UniversalDeserializer()

	obj, _, err := decoder.Decode([]byte(yamlString), nil, nil)
	if err != nil {
		return nil, err
	}

	return obj, nil
}
