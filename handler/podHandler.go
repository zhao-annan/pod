package handler

import (
	"context"
	"github.com/zhao-annan/common"
	"github.com/zhao-annan/pod/domain/model"
	"github.com/zhao-annan/pod/domain/service"
	"github.com/zhao-annan/pod/proto/pod"
	"strconv"
)

type PodHandler struct {

	//注意这里的类型是IPodDataService接口类型

	PodDataService service.IPodDataService
}

//添加创建POD

func (e *PodHandler) AddPod(ctx context.Context, info *pod.PodInfo, rsp *pod.Response) error {

	common.Info("添加pod")

	podModel := &model.Pod{}

	err := common.SwapTo(info, podModel)

	if err != nil {

		common.Error(err)

		rsp.Msg = err.Error()

		return err

	}
	if err := e.PodDataService.CreateToK8s(info); err != nil {
		common.Error(err)
		rsp.Msg = err.Error()
		return err
	} else {
		//操作数据库写入数据库

		podID, err := e.PodDataService.AddPod(podModel)
		if err != nil {

			common.Error(err)
			rsp.Msg = err.Error()
			return err
		}

		common.Info("Pod 添加成功数据库ID号为:" + strconv.FormatInt(podID, 10))
		rsp.Msg = "Pod 添加成功数据库ID号为:" + strconv.FormatInt(podID, 10)

	}

	return nil

}

//删除POD

func (e *PodHandler) DeletePod(ctx context.Context, req *pod.PodId,
	rsp *pod.Response) error {

	//先查找数据
	podModel, err := e.PodDataService.FindPodByID(req.Id)

	if err != nil {

		common.Error(err)
		return err
	}
	if err := e.PodDataService.DeleteFromK8s(podModel); err != nil {
		common.Error(err)
		return err
	}
	return nil
}

//更新指定的pod
func (e *PodHandler) UpdatePod(ctx context.Context, req *pod.PodInfo,
	rsp *pod.Response) error {
	err := e.PodDataService.UpdateToK8s(req)

	if err != nil {
		common.Error(err)
		return err
	}
	//查询数据库中的数据的pod
	podModel, err := e.PodDataService.FindPodByID(req.Id)

	if err != nil {
		common.Error(err)
		return err
	}
	err = common.SwapTo(req, podModel)
	if err != nil {
		common.Error(err)
		return err
	}
	e.PodDataService.UpdatePod(podModel)
	return nil
}

//查询单个信息
func (e *PodHandler) FindPodByID(ctx context.Context, req *pod.PodId,
	rsp *pod.PodInfo) error {
	//先从数据库中查到数据
	podModel, err := e.PodDataService.FindPodByID(req.Id)
	if err != nil {
		common.Error(err)
		return err
	}
	err = common.SwapTo(podModel, rsp)

	if err != nil {
		common.Error(err)
		return err
	}

	return nil

}

func (e *PodHandler) FindAllPod(ctx context.Context, req *pod.FindAll,
	rsp *pod.AllPod) error {

	//查询所有的pod
	allPod, err := e.PodDataService.FindAllPod()

	if err != nil {
		common.Error(err)
		return err
	}

	//整理格式

	for _, v := range allPod {

		podInfo := &pod.PodInfo{}

		err := common.SwapTo(v, podInfo)
		if err != nil {
			common.Error(err)
			return err
		}
		rsp.PodInfo = append(rsp.PodInfo, podInfo)
	}

	return nil

}
