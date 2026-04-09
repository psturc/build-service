package imagecontroller

import (
	"context"

	"github.com/konflux-ci/image-controller/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	rclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func (i *ImageController) CreateImageRepositoryCR(name, namespace, applicationName, componentName string) (*v1alpha1.ImageRepository, error) {
	imageRepository := &v1alpha1.ImageRepository{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"appstudio.redhat.com/application": applicationName,
				"appstudio.redhat.com/component":   componentName,
			},
		},
	}
	err := i.KubeRest().Create(context.Background(), imageRepository)
	if err != nil {
		return nil, err
	}
	return imageRepository, nil
}

func (i *ImageController) GetImageRepositoryCR(name, namespace string) (*v1alpha1.ImageRepository, error) {
	namespacedName := types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}

	imageRepository := v1alpha1.ImageRepository{}

	err := i.KubeRest().Get(context.Background(), namespacedName, &imageRepository)
	if err != nil {
		return nil, err
	}
	return &imageRepository, nil
}

func (i *ImageController) ChangeVisibilityToPrivate(namespace, applicationName, componentName string) (*v1alpha1.ImageRepository, error) {
	imageRepositoryList := &v1alpha1.ImageRepositoryList{}
	imageRepoLabels := map[string]string{"appstudio.redhat.com/component": componentName}
	err := i.KubeRest().List(context.Background(), imageRepositoryList, &rclient.ListOptions{LabelSelector: labels.SelectorFromSet(imageRepoLabels), Namespace: namespace})
	if err != nil {
		return nil, err
	}
	imageRepository := &imageRepositoryList.Items[0]
	imageRepository.Spec.Image.Visibility = "private"

	err = i.KubeRest().Update(context.Background(), imageRepository)
	if err != nil {
		return nil, err
	}
	return imageRepository, nil
}

func (i *ImageController) GetImageName(namespace, componentName string) (string, error) {
	imageRepositoryList := &v1alpha1.ImageRepositoryList{}
	imageRepoLabels := map[string]string{"appstudio.redhat.com/component": componentName}
	err := i.KubeRest().List(context.Background(), imageRepositoryList, &rclient.ListOptions{LabelSelector: labels.SelectorFromSet(imageRepoLabels), Namespace: namespace})
	if err != nil {
		return "", err
	}
	return imageRepositoryList.Items[0].Spec.Image.Name, err
}

func (i *ImageController) GetRobotAccounts(namespace, componentName string) (string, string, error) {
	imageRepositoryList := &v1alpha1.ImageRepositoryList{}
	imageRepoLabels := map[string]string{"appstudio.redhat.com/component": componentName}
	err := i.KubeRest().List(context.Background(), imageRepositoryList, &rclient.ListOptions{LabelSelector: labels.SelectorFromSet(imageRepoLabels), Namespace: namespace})
	if err != nil {
		return "", "", err
	}
	return imageRepositoryList.Items[0].Status.Credentials.PullRobotAccountName, imageRepositoryList.Items[0].Status.Credentials.PushRobotAccountName, nil
}

func (i *ImageController) IsVisibilityPublic(namespace, componentName string) (bool, error) {
	imageRepositoryList := &v1alpha1.ImageRepositoryList{}
	imageRepoLabels := map[string]string{"appstudio.redhat.com/component": componentName}
	err := i.KubeRest().List(context.Background(), imageRepositoryList, &rclient.ListOptions{LabelSelector: labels.SelectorFromSet(imageRepoLabels), Namespace: namespace})
	if err != nil {
		return false, err
	}
	return imageRepositoryList.Items[0].Spec.Image.Visibility == "public", nil
}
