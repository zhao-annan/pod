package service

import (
	"context"
	"errors"
	"github.com/zhao-annan/common"
	"github.com/zhao-annan/pod/domain/model"
	"github.com/zhao-annan/pod/domain/repository"
	"github.com/zhao-annan/pod/proto/pod"
	v1 "k8s.io/api/apps/v1"
	v13 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"strconv"
)

type IPodDataService interface {
	AddPod(*model.Pod) (int64, error)
	DeletePod(int64) error
	UpdatePod(*model.Pod) error
	FindPodByID(int64) (*model.Pod, error)
	FindAllPod() ([]model.Pod, error)
	CreateToK8s(*pod.PodInfo) error
	DeleteFromK8s(*model.Pod) error
	UpdateToK8s(*pod.PodInfo) error
}

func NewPodDataService(podRepository repository.IPodRepository, clientSet *kubernetes.Clientset) IPodDataService {

	return &PodDataService{
		PodRepository: podRepository,
		k8sClientSet:  clientSet,
		deployment:    &v1.Deployment{},
	}

}

type PodDataService struct {
	PodRepository repository.IPodRepository
	k8sClientSet  *kubernetes.Clientset
	deployment    *v1.Deployment
}

//func (u *PodDataService) UpdateToK8s(info *pod.PodInfo) error {
//	//TODO implement me
//	panic("implement me")
//}

//更新deployment
func (u *PodDataService) UpdateToK8s(podInfo *pod.PodInfo) (err error) {
	//func (u *PodDataService) UpdateToK8s(podInfo *pod.PodInfo) (err error) {
	u.SetDeployment(podInfo)
	if _, err = u.k8sClientSet.AppsV1().Deployments(podInfo.PodNamespace).Get(
		context.TODO(), podInfo.PodName, v12.GetOptions{}); err != nil {

		common.Error(err)
		return errors.New("Pod" + podInfo.PodName + "不存在请先创建")
	} else {

		//如果存在执行更新操作

		if _, err = u.k8sClientSet.AppsV1().Deployments(podInfo.PodNamespace).
			Update(context.TODO(), u.deployment, v12.UpdateOptions{}); err != nil {

			common.Error(err)
			return err
		}

		common.Info(podInfo.PodName + "更新成功")
		return nil

	}

}

//创建pod到k8s中
func (u *PodDataService) CreateToK8s(podInfo *pod.PodInfo) (err error) {

	u.SetDeployment(podInfo)
	//先判断有没有这个pod，如果有了 就不需要创建，没有再创建这个Pod
	if _, err = u.k8sClientSet.AppsV1().Deployments(podInfo.PodNamespace).Get(
		context.TODO(), podInfo.PodName, v12.GetOptions{}); err != nil {
		if _, err = u.k8sClientSet.AppsV1().Deployments(podInfo.PodNamespace).
			Create(context.TODO(), u.deployment, v12.CreateOptions{}); err != nil {
			common.Error(err)
			return err
		}
		common.Info("创建成功")
		return nil
	} else {
		//可以写自己的业务逻辑
		common.Error("Pod" + podInfo.PodName + "已经存在")
		return errors.New("Pod" + podInfo.PodName + "已经存在")
	}

	//common.Info("创建成功")
	//return nil

}

//删除操作 删除pod
func (u *PodDataService) DeleteFromK8s(pod *model.Pod) (err error) {
	if err = u.k8sClientSet.AppsV1().Deployments(pod.PodNameSpace).Delete(
		context.TODO(), pod.PodName, v12.DeleteOptions{}); err != nil {

		common.Error(err)

		//写自己的业务逻辑

		return err
	} else {
		if err := u.DeletePod(pod.ID); err != nil {
			common.Error(err)
			return err
		}
		common.Info("删除Pod ID:" + strconv.FormatInt(pod.ID, 10) + "成功!")
	}
	return

}

func (u *PodDataService) SetDeployment(podInfo *pod.PodInfo) {

	deployment := &v1.Deployment{}
	deployment.TypeMeta = v12.TypeMeta{
		Kind:       "deployment",
		APIVersion: "v1",
	}
	//这里的ObjectMeta对应的就是metadata，也就是元对象对应的是元数据
	deployment.ObjectMeta = v12.ObjectMeta{
		Name: podInfo.PodName,

		Namespace: podInfo.PodNamespace,

		Labels: map[string]string{

			"app-name": podInfo.PodName,
			"author":   "zhao-annan",
		},
	}
	deployment.Name = podInfo.PodName
	deployment.Spec = v1.DeploymentSpec{

		//副本个数
		Replicas: &podInfo.PodReplicas,
		//用来查找关联的Pod，所有标签都匹配才行
		Selector: &v12.LabelSelector{
			MatchLabels: map[string]string{
				"app-name": podInfo.PodName,
			},
			MatchExpressions: nil,
		},
		//定义Pod相关数据
		Template: v13.PodTemplateSpec{
			ObjectMeta: v12.ObjectMeta{
				Labels: map[string]string{
					"app-name": podInfo.PodName,
				},
			},
			Spec: v13.PodSpec{

				Containers: []v13.Container{
					{
						Name:  podInfo.PodName,
						Image: podInfo.PodImage,
						Ports: u.getContainerPort(podInfo),
						Env:   u.getEnv(podInfo),

						Resources:       u.getResources(podInfo),
						ImagePullPolicy: u.getImagePullPolicy(podInfo),
					},
				},
			},
		},
		Strategy:                v1.DeploymentStrategy{},
		MinReadySeconds:         0,
		RevisionHistoryLimit:    nil,
		Paused:                  false,
		ProgressDeadlineSeconds: nil,
	}

	u.deployment = deployment

}

func (u *PodDataService) getContainerPort(podInfo *pod.PodInfo) (
	containerPort []v13.ContainerPort) {

	for _, v := range podInfo.PodPort {

		containerPort = append(containerPort, v13.ContainerPort{
			Name:          "port-" + strconv.FormatInt(int64(v.ContainerPort), 10),
			ContainerPort: v.ContainerPort,
			Protocol:      u.getProtocol(v.Protocol),
		})

	}

	return

}

func (u *PodDataService) getProtocol(protocol string) v13.Protocol {

	switch protocol {
	case "TCP":
		return "TCP"
	case "UDP":
		return "UDP"
	case "SCTP":
		return "SCTP"
	default:
		return "TCP"
	}

}

func (u *PodDataService) getEnv(podInfo *pod.PodInfo) (envVar []v13.EnvVar) {

	for _, v := range podInfo.PodEnv {

		envVar = append(envVar, v13.EnvVar{
			Name:      v.EnvKey,
			Value:     v.EnvValue,
			ValueFrom: nil,
		})
	}
	return
}

func (u *PodDataService) getResources(podInfo *pod.PodInfo) (source v13.
	ResourceRequirements) {
	//最大能够使用多少资源
	source.Limits = v13.ResourceList{
		"cpu": resource.MustParse(strconv.FormatFloat(float64(podInfo.PodCpuMax),
			'f', 6, 64)),
		"memory": resource.MustParse(strconv.FormatFloat(float64(podInfo.PodMemoryMax),
			'f', 6, 64)),
	}
	//满足最少使用的资源量

	//@TODO  自己实现动态设置
	source.Requests = v13.ResourceList{
		"cpu": resource.MustParse(strconv.FormatFloat(float64(podInfo.PodCpuMax),
			'f', 6, 64)),
		"memory": resource.MustParse(strconv.FormatFloat(float64(podInfo.PodMemoryMax),
			'f', 6, 64)),
	}
	return

}

func (u *PodDataService) getImagePullPolicy(podInfo *pod.PodInfo) v13.PullPolicy {

	switch podInfo.PodPullPolicy {
	case "Always":
		return "Always"
	case "Never":
		return "Never"
	case "IfNotPresent":
		return "IfNotPresent"
	default:
		return "Always"

	}

}

//插入操作
func (u *PodDataService) AddPod(pod2 *model.Pod) (int64, error) {
	return u.PodRepository.CreatePod(pod2)

}

//删除操作
func (u *PodDataService) DeletePod(podID int64) error {
	return u.PodRepository.DeletePodByID(podID)

}

//更新
func (u *PodDataService) UpdatePod(pod2 *model.Pod) error {
	return u.PodRepository.UpdatePod(pod2)

}

//单个查找
func (u *PodDataService) FindPodByID(podID int64) (*model.Pod, error) {

	return u.PodRepository.FindPodByID(podID)
}

//查找所有
func (u *PodDataService) FindAllPod() ([]model.Pod, error) {
	return u.PodRepository.FindAll()
}
