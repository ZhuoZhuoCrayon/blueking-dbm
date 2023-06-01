# -*- coding: utf-8 -*-
"""
TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-DB管理系统(BlueKing-BK-DBM) available.
Copyright (C) 2017-2023 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
You may obtain a copy of the License at https://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
specific language governing permissions and limitations under the License.
"""
import logging
from typing import List

from django.utils.translation import ugettext as _
from pipeline.component_framework.component import Component
from pipeline.core.flow.activity import Service

from backend.flow.consts import DnsOpType
from backend.flow.plugins.components.collections.common.base_service import BaseService
from backend.flow.utils.dns_manage import DnsManage

logger = logging.getLogger("flow")


class KafkaDnsManageService(BaseService):
    """
    定义mysql集群域名管理的活动节点,目前只支持添加域名、删除域名
    """

    def __get_exec_ips(self, global_data) -> list:
        """
        获取需要执行的ip list
        """
        exec_ips = [broker["ip"] for broker in global_data["nodes"]["broker"]]

        if not exec_ips:
            self.log_error(_("该节点获取到执行ip信息为空，请联系系统管理员"))

        return exec_ips

    def __get_ips(self, global_data) -> (list, list):
        """
        获取需要执行的ip list
        """
        old_ips = [broker["ip"] for broker in global_data["old_nodes"]["broker"]]
        new_ips = [broker["ip"] for broker in global_data["new_nodes"]["broker"]]
        if not old_ips or not new_ips:
            self.log_error(_("该节点获取到执行ip信息为空，请联系系统管理员"))

        return old_ips, new_ips

    def _execute(self, data, parent_data) -> bool:
        kwargs = data.get_one_of_inputs("kwargs")
        global_data = data.get_one_of_inputs("global_data")

        # 传入调用结果
        dns_op_type = kwargs["dns_op_type"]
        dns_manage = DnsManage(bk_biz_id=global_data["bk_biz_id"], bk_cloud_id=kwargs["bk_cloud_id"])
        if dns_op_type == DnsOpType.CREATE:
            exec_ips = self.__get_exec_ips(global_data=global_data)
            if not exec_ips:
                return False
            add_instance_list = [f"{ip}#{kwargs['dns_op_exec_port']}" for ip in exec_ips]
            # 添加域名映射
            result = dns_manage.create_domain(
                instance_list=add_instance_list, add_domain_name=kwargs["add_domain_name"]
            )
        elif dns_op_type == DnsOpType.CLUSTER_DELETE:
            # 清理域名
            result = dns_manage.delete_domain(cluster_id=global_data["cluster_id"])
        elif dns_op_type == DnsOpType.RECYCLE_RECORD:
            # 回收实例对应的域名记录，适配实例缩容场景
            exec_ips = self.__get_exec_ips(global_data=global_data)
            if not exec_ips:
                self.log_error(_("该节点获取到执行ip信息为空，请联系系统管理员"))
                return False

            delete_instance_list = [f"{ip}#{kwargs['dns_op_exec_port']}" for ip in exec_ips]
            result = dns_manage.remove_domain_ip(domain=global_data["domain"], del_instance_list=delete_instance_list)
        elif dns_op_type == DnsOpType.UPDATE:
            old_ips, new_ips = self.__get_ips(global_data=global_data)
            if not old_ips or not new_ips:
                return False
            old_instance_list = [f"{ip}#{kwargs['dns_op_exec_port']}" for ip in old_ips]
            new_instance_list = [f"{ip}#{kwargs['dns_op_exec_port']}" for ip in new_ips]
            # 更新域名
            result = dns_manage.batch_update_domain(
                old_instance_list=old_instance_list,
                new_instance_list=new_instance_list,
                update_domain_name=kwargs["add_domain_name"],
            )
        else:
            logger.error(_("无法适配到传入的域名处理类型,请联系系统管理员:{}").format(dns_op_type))
            return False
        return result

    def inputs_format(self) -> List:
        return [
            Service.InputItem(name="kwargs", key="kwargs", type="dict", required=True),
            Service.InputItem(name="global_data", key="global_data", type="dict", required=True),
        ]


class KafkaDnsManageComponent(Component):
    name = __name__
    code = "kafka_dns_manage"
    bound_service = KafkaDnsManageService