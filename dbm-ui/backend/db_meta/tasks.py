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
import datetime
import json
import logging

from celery import shared_task

from .. import env
from ..components import CCApi
from ..dbm_init.constants import CC_APP_ABBR_ATTR, CC_HOST_DBM_ATTR
from .models import AppCache, Machine
from .models.cluster_monitor import SyncFailedMachine

logger = logging.getLogger("celery")


@shared_task
def update_app_cache():
    """缓存空闲机拓扑"""

    now = datetime.datetime.now()
    logger.warning("[db_meta] start update app cache start: %s", now)

    bizs = CCApi.search_business().get("info", [])

    updated_hosts, created_hosts = [], []
    for biz in bizs:
        try:
            logger.warning("[db_meta] sync app : %s", biz["bk_biz_id"])
            bk_app_abbr = biz.get(env.BK_APP_ABBR, "")
            db_app_abbr = biz.get(CC_APP_ABBR_ATTR, "")

            # 目标环境中存在bk_app_abbr，则同步过来
            if env.BK_APP_ABBR and env.BK_APP_ABBR != CC_APP_ABBR_ATTR:
                # db_app_abbr为空才同步
                if not db_app_abbr and db_app_abbr != bk_app_abbr:
                    CCApi.update_business(
                        {"bk_biz_id": biz["bk_biz_id"], "data": {"db_app_abbr": bk_app_abbr}}, use_admin=True
                    )
                    db_app_abbr = bk_app_abbr

            obj, created = AppCache.objects.update_or_create(
                defaults={
                    "db_app_abbr": db_app_abbr,
                    "bk_biz_name": biz["bk_biz_name"],
                    "language": biz["language"],
                    "time_zone": biz["time_zone"],
                    "bk_biz_maintainer": biz["bk_biz_maintainer"],
                },
                bk_biz_id=biz["bk_biz_id"],
            )

            if created:
                created_hosts.append(obj.bk_biz_id)
            else:
                updated_hosts.append(obj.bk_biz_id)
        except Exception as e:  # pylint: disable=wildcard-import
            logger.error("[db_meta] cache app error: %s (%s)", biz, e)

    logger.warning(
        "[db_meta] finish update app cache end: %s, create_cnt: %s, update_cnt: %s",
        datetime.datetime.now() - now,
        len(created_hosts),
        len(updated_hosts),
    )


@shared_task
def update_host_dbmeta(bk_biz_id=None):
    """更新集群主机的dbm_meta属性"""

    now = datetime.datetime.now()
    logger.info("[update_host_dbmeta] start update begin: %s", now)

    # 0为默认的无效machine
    failed_host_ids = list(SyncFailedMachine.objects.all().values_list("bk_host_id", flat=True)) + [0]

    # 支持业务级别的更新
    machines = Machine.objects.all()
    if bk_biz_id:
        machines = machines.filter(bk_biz_id=bk_biz_id)

    machines = machines.exclude(
        bk_host_id__in=failed_host_ids,
    ).order_by("-create_at")

    # 批量更新接口限制最多500条，这里取456条
    STEP = 456
    updated_hosts, failed_updates = [], []
    machine_count = machines.count()
    for step in range(machine_count // STEP + 1):
        updates = []
        for machine in machines[step * STEP : (step + 1) * STEP]:
            updates.append(
                {"properties": {CC_HOST_DBM_ATTR: json.dumps(machine.dbm_meta)}, "bk_host_id": machine.bk_host_id}
            )
        updated_hosts.extend(updates)

        try:
            CCApi.batch_update_host({"update": updates}, use_admin=True)
        except Exception as e:  # pylint: disable=wildcard-import
            failed_updates.extend(updates)
            logger.error("[update_host_dbmeta] batch update exception: %s (%s)", updates, e)

    # 容错处理：逐个更新，避免批量更新误伤有效ip
    for fail_update in failed_updates:
        try:
            CCApi.update_host(
                {"bk_host_id": fail_update["bk_host_id"], "data": fail_update["properties"]}, use_admin=True
            )
        except Exception as e:  # pylint: disable=wildcard-import
            # 记录异常ip，下次任务直接排除掉，尽量走批量更新
            SyncFailedMachine.objects.get_or_create(bk_host_id=fail_update["bk_host_id"], error=str(e))
            logger.error("[update_host_dbmeta] single update error: %s (%s)", fail_update, e)

    logger.info(
        "[update_host_dbmeta] finish update end: %s, update_cnt: %s",
        datetime.datetime.now() - now,
        len(updated_hosts),
    )
