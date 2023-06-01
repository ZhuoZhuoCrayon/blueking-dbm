// Package jobmanager 原子任务工厂类 与 管理类
package jobmanager

import (
	"dbm-services/redis/db-tools/dbactuator/pkg/atomjobs/atommongodb"
	"dbm-services/redis/db-tools/dbactuator/pkg/atomjobs/atomproxy"
	"dbm-services/redis/db-tools/dbactuator/pkg/atomjobs/atomredis"
	"dbm-services/redis/db-tools/dbactuator/pkg/atomjobs/atomsys"
	"dbm-services/redis/db-tools/dbactuator/pkg/jobruntime"
	"dbm-services/redis/db-tools/dbactuator/pkg/util"
	"fmt"
	"log"
	"runtime/debug"
	"strings"
	"sync"
	"time"
)

// AtomJobCreatorFunc 原子任务创建接口
type AtomJobCreatorFunc func() jobruntime.JobRunner

// JobGenericManager 原子任务管理者
type JobGenericManager struct {
	Runners       []jobruntime.JobRunner `json:"runners"`
	atomJobMapper map[string]AtomJobCreatorFunc
	once          sync.Once
	runtime       *jobruntime.JobGenericRuntime
}

// NewJobGenericManager new
func NewJobGenericManager(uid, rootID, nodeID, versionID, payload, payloadFormat, atomJobs, baseDir string) (
	ret *JobGenericManager, err error) {
	runtime, err := jobruntime.NewJobGenericRuntime(uid, rootID, nodeID, versionID,
		payload, payloadFormat, atomJobs, baseDir)
	if err != nil {
		log.Panicf(err.Error())
	}
	ret = &JobGenericManager{
		runtime: runtime,
	}
	return
}

// LoadAtomJobs 加载子任务
func (m *JobGenericManager) LoadAtomJobs() (err error) {
	defer func() {
		// err最后输出到标准错误
		if err != nil {
			m.runtime.PrintToStderr(err.Error())
		}
	}()
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%s", (debug.Stack()))
		}
	}()
	m.runtime.AtomJobList = strings.TrimSpace(m.runtime.AtomJobList)
	if m.runtime.AtomJobList == "" {
		err = fmt.Errorf("atomJobList(%s) cannot be empty", m.runtime.AtomJobList)
		m.runtime.Logger.Error(err.Error())
		return
	}
	jobList := strings.Split(m.runtime.AtomJobList, ",")
	for _, atomName := range jobList {
		atomName = strings.TrimSpace(atomName)
		if atomName == "" {
			continue
		}
		atom := m.GetAtomJobInstance(atomName)
		if atom == nil {
			err = fmt.Errorf("atomJob(%s) not found", atomName)
			m.runtime.Logger.Error(err.Error())
			return
		}
		m.Runners = append(m.Runners, atom)
		m.runtime.Logger.Info(fmt.Sprintf("atomJob:%s instance load success", atomName))
	}
	return
}

// RunAtomJobs 顺序执行原子任务
func (m *JobGenericManager) RunAtomJobs() (err error) {
	defer func() {
		// err最后输出到标准错误
		if err != nil {
			m.runtime.PrintToStderr(err.Error() + "\n")
		}
	}()
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%s", string(debug.Stack()))
		}
	}()

	m.runtime.StartHeartbeat(10 * time.Second)

	defer m.runtime.StopHeartbeat()

	for _, runner := range m.Runners {
		name := util.GetTypeName(runner)
		m.runtime.Logger.Info(fmt.Sprintf("begin to run %s init", name))
		if err = runner.Init(m.runtime); err != nil {
			return
		}
		m.runtime.Logger.Info(fmt.Sprintf("begin to run %s", name))
		err = runner.Run()
		if err != nil {
			m.runtime.Logger.Info(fmt.Sprintf("runner %s run failed,err:%s", name, err))
			// err = runner.Rollback()
			// if err != nil {
			// 	err = fmt.Errorf("runner %s rollback failed,err:%+v", name, err)
			// 	m.runtime.Logger.Error(err.Error())
			// 	return
			// }
			// m.runtime.Logger.Info(fmt.Sprintf("runner %s rollback success!!!", name))
			return
		}
		m.runtime.Logger.Info(fmt.Sprintf("finished run %s", name))
	}
	m.runtime.Logger.Info(fmt.Sprintf("run all atomJobList:%s success", m.runtime.AtomJobList))

	m.runtime.OutputPipeContextData()
	return
}

func (m *JobGenericManager) atomjobsMapperLoading() {
	m.once.Do(func() {
		m.atomJobMapper = make(map[string]AtomJobCreatorFunc)
		m.atomJobMapper[atomsys.NewSysInit().Name()] = atomsys.NewSysInit

		// redis atom jobs
		m.atomJobMapper[atomredis.NewRedisInstall().Name()] = atomredis.NewRedisInstall
		m.atomJobMapper[atomredis.NewRedisReplicaOf().Name()] = atomredis.NewRedisReplicaOf
		m.atomJobMapper[atomredis.NewRedisReplicaBatch().Name()] = atomredis.NewRedisReplicaBatch
		m.atomJobMapper[atomredis.NewClusterMeetSlotsAssign().Name()] = atomredis.NewClusterMeetSlotsAssign
		m.atomJobMapper[atomproxy.NewTwemproxyInstall().Name()] = atomproxy.NewTwemproxyInstall
		m.atomJobMapper[atomredis.NewRedisBackup().Name()] = atomredis.NewRedisBackup
		m.atomJobMapper[atomredis.NewTendisKeysPattern().Name()] = atomredis.NewTendisKeysPattern
		m.atomJobMapper[atomredis.NewTendisKeysPatternDelete().Name()] = atomredis.NewTendisKeysPatternDelete
		m.atomJobMapper[atomredis.NewTendisKeysFilesDelete().Name()] = atomredis.NewTendisKeysFilesDelete
		m.atomJobMapper[atomproxy.NewPredixyInstall().Name()] = atomproxy.NewPredixyInstall
		m.atomJobMapper[atomredis.NewTendisssdDrRestore().Name()] = atomredis.NewTendisssdDrRestore
		m.atomJobMapper[atomproxy.NewTwemproxyOperate().Name()] = atomproxy.NewTwemproxyOperate
		m.atomJobMapper[atomproxy.NewPredixyOperate().Name()] = atomproxy.NewPredixyOperate
		m.atomJobMapper[atomredis.NewRedisShutdown().Name()] = atomredis.NewRedisShutdown
		m.atomJobMapper[atomredis.NewRedisFlushData().Name()] = atomredis.NewRedisFlushData
		m.atomJobMapper[atomsys.NewRedisCapturer().Name()] = atomsys.NewRedisCapturer
		m.atomJobMapper[atomredis.NewRedisSwitch().Name()] = atomredis.NewRedisSwitch
		m.atomJobMapper[atomredis.NewBkDbmonInstall().Name()] = atomredis.NewBkDbmonInstall
		m.atomJobMapper[atomredis.NewTendisPlusMigrateSlots().Name()] = atomredis.NewTendisPlusMigrateSlots
		m.atomJobMapper[atomredis.NewRedisDtsDataCheck().Name()] = atomredis.NewRedisDtsDataCheck
		m.atomJobMapper[atomredis.NewRedisDtsDataRepaire().Name()] = atomredis.NewRedisDtsDataRepaire
		// scene needs.
		m.atomJobMapper[atomproxy.NewTwemproxySceneCheckBackends().Name()] = atomproxy.NewTwemproxySceneCheckBackends
		m.atomJobMapper[atomredis.NewRedisSceneSyncCheck().Name()] = atomredis.NewRedisSceneSyncCheck
		m.atomJobMapper[atomredis.NewRedisSceneKillDeadConn().Name()] = atomredis.NewRedisSceneKillDeadConn
		m.atomJobMapper[atomredis.NewRedisSceneSyncPrams().Name()] = atomredis.NewRedisSceneSyncPrams

		// mongo atom jobs
		m.atomJobMapper[atommongodb.NewMongoDBInstall().Name()] = atommongodb.NewMongoDBInstall
		m.atomJobMapper[atommongodb.NewMongoSInstall().Name()] = atommongodb.NewMongoSInstall
		m.atomJobMapper[atommongodb.NewInitiateReplicaset().Name()] = atommongodb.NewInitiateReplicaset
		m.atomJobMapper[atommongodb.NewAddShardToCluster().Name()] = atommongodb.NewAddShardToCluster
		m.atomJobMapper[atommongodb.NewAddUser().Name()] = atommongodb.NewAddUser
		m.atomJobMapper[atommongodb.NewDelUser().Name()] = atommongodb.NewDelUser
		m.atomJobMapper[atommongodb.NewMongoDReplace().Name()] = atommongodb.NewMongoDReplace
		m.atomJobMapper[atommongodb.NewMongoRestart().Name()] = atommongodb.NewMongoRestart
		m.atomJobMapper[atommongodb.NewStepDown().Name()] = atommongodb.NewStepDown
		m.atomJobMapper[atommongodb.NewBalancer().Name()] = atommongodb.NewBalancer
		m.atomJobMapper[atommongodb.NewDeInstall().Name()] = atommongodb.NewDeInstall
		m.atomJobMapper[atommongodb.NewExecScript().Name()] = atommongodb.NewExecScript
		m.atomJobMapper[atommongodb.NewSetProfiler().Name()] = atommongodb.NewSetProfiler
		m.atomJobMapper[atomsys.NewOsMongoInit().Name()] = atomsys.NewOsMongoInit
	})
}

// SupportAtomJobs 返回支持的atomJob列表
func (m *JobGenericManager) SupportAtomJobs() []string {
	m.atomjobsMapperLoading()
	atomJobs := make([]string, 0, len(m.atomJobMapper))
	for k := range m.atomJobMapper {
		atomJobs = append(atomJobs, k)
	}
	return atomJobs
}

// GetAtomJobInstance 根据atomJobName,从m.atomJobMapper中获取其creator函数,执行creator函数
func (m *JobGenericManager) GetAtomJobInstance(atomJob string) jobruntime.JobRunner {
	m.atomjobsMapperLoading()
	creator, ok := m.atomJobMapper[strings.ToLower(atomJob)]
	if ok {
		return creator()
	}
	return nil
}