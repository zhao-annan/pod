package repository

import (
	"git.imooc.com/zhao-annan/pod/domain/model"
	"github.com/jinzhu/gorm"
)

//创建需要实现的接口
type IPodRepository interface {

	//初始化表

	InitTable() error
	//根据ID查找数据

	FindPodByID(int64) (*model.Pod, error)

	//创建一条pod 数据
	CreatePod(*model.Pod) (int64, error)

	//根据ID删除一条Pod数据

	DeletePodByID(int64) error

	//修改一条数据
	UpdatePod(*model.Pod) error
	//查找Pod所有数据

	FindAll() ([]model.Pod, error)
}

//创建PodRepository
func NewPodRepository(db *gorm.DB) IPodRepository {
	return &PodRepository{mysqlDb: db}
}

type PodRepository struct {
	mysqlDb *gorm.DB
}

//初始化表

func (u *PodRepository) InitTable() error {
	return u.mysqlDb.CreateTable(&model.Pod{}, &model.PodEnv{},
		&model.PodPort{}).Error
}

func (u *PodRepository) FindPodByID(podID int64) (pod *model.Pod, err error) {
	pod = &model.Pod{}
	return pod, u.mysqlDb.Preload("PodEnv").Preload(
		"PodPort").First(pod, podID).Error
}

func (u *PodRepository) CreatePod(pod *model.Pod) (int64, error) {

	return pod.ID, u.mysqlDb.Create(pod).Error
}

//删除  根据ID删除Pod信息
func (u *PodRepository) DeletePodByID(podID int64) error {

	tx := u.mysqlDb.Begin()
	//遇到问题回滚
	defer func() {

		if r := recover(); r != nil {
			tx.Rollback()
		}

	}()
	if tx.Error != nil {
		return tx.Error
	}
	//彻底删除  pod信息

	if err := u.mysqlDb.Where("id=?", podID).Delete(&model.Pod{}).
		Error; err != nil {
		tx.Rollback()
		return err

	}

	//彻底删除  podenv信息

	if err := u.mysqlDb.Where("pod_id = ?",
		podID).Delete(&model.PodEnv{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := u.mysqlDb.Where("pod_id = ?",
		podID).Delete(&model.PodPort{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	//提交事务
	return tx.Commit().Error

}

//更新Pod信息

func (u *PodRepository) UpdatePod(pod *model.Pod) error {
	return u.mysqlDb.Model(pod).Update(pod).Error

}

//获取所有结果集
func (u *PodRepository) FindAll() (podAll []model.Pod, err error) {

	return podAll, u.mysqlDb.Find(&podAll).Error
}
