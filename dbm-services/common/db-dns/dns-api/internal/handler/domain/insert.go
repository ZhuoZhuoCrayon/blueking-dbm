package domain

import (
	"bk-dnsapi/internal/domain/entity"
	"bk-dnsapi/internal/domain/repo/domain"
	"bk-dnsapi/pkg/tools"
	"encoding/json"
	"fmt"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-mesh/openlogging"
)

// DnsBasePutReqParam TODO
type DnsBasePutReqParam struct {
	// Appid 	int64		`json:"appid"`
	App       string `json:"app,required"`
	BkCloudId int64  `json:"bk_cloud_id"`
	Domains   []struct {
		DomainName string   `json:"domain_name"`
		Instances  []string `json:"instances,required"`
		Manager    string   `json:"manager"`
		Remark     string   `json:"remark"`
		DomainType string   `json:"domain_type"`
	} `json:"domains"`
}

// AddDns TODO
func (h *Handler) AddDns(c *gin.Context) {
	defer func() {
		if r := recover(); r != nil {
			openlogging.Error(fmt.Sprintf("panic error:%v,stack:%s", r, string(debug.Stack())))
			SendResponse(c,
				fmt.Errorf("panic error:%v", r),
				Data{})
		}
	}()

	var addParam DnsBasePutReqParam
	err := c.BindJSON(&addParam)
	if err != nil {
		SendResponse(c, err, Data{})
		return
	}

	openlogging.Info(fmt.Sprintf("add dns begin, param [%+v]", addParam))

	// TODO check
	// check app exists
	var errMsg string
	var dnsBaseList []*entity.TbDnsBase
	for i := 0; i < len(addParam.Domains); i++ {
		domain := addParam.Domains[i]
		if domain.DomainName, err = tools.CheckDomain(domain.DomainName); err != nil {
			errMsg += err.Error() + "\r\n"
			continue
		}
		for j := 0; j < len(domain.Instances); j++ {
			ins := strings.TrimSpace(domain.Instances[j])
			// 支持ip格式，默认端口为0
			if !strings.Contains(ins, "#") {
				ins += "#0"
			}
			ip, port, err := tools.GetIpPortByIns(ins)
			if err != nil {
				errMsg += err.Error() + "\r\n"
				continue
			}
			// ip, _ = tools.CheckIp(ip)
			if domain.Manager == "" {
				domain.Manager = "DBAManager"
			}

			t := &entity.TbDnsBase{
				Uid:            0,
				App:            addParam.App,
				DomainName:     domain.DomainName,
				Ip:             ip,
				Port:           port,
				StartTime:      time.Now(),
				LastChangeTime: time.Now(),
				Manager:        domain.Manager,
				Remark:         domain.Remark,
				Status:         "1",
				BkCloudId:      addParam.BkCloudId,
			}

			dnsBaseList = append(dnsBaseList, t)
		}
	}

	if errMsg != "" {
		SendResponse(c, fmt.Errorf(errMsg), Data{})
		return
	}
	info, _ := json.Marshal(dnsBaseList)
	openlogging.Info(fmt.Sprintf("add insert begin exec, param [%+v]", string(info)))

	rowsAffected, err := domain.DnsDomainResource().Insert(dnsBaseList)
	SendResponse(c, err, Data{RowsNum: rowsAffected})
}
