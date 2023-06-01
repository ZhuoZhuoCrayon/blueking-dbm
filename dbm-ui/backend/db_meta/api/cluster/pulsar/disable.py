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

from django.db import transaction

from backend.db_meta.enums import ClusterPhase
from backend.db_meta.models import Cluster

logger = logging.getLogger("root")


@transaction.atomic
def disable(cluster_id: int):
    """
    禁用Pulsar集群
    """

    cluster = Cluster.objects.get(id=cluster_id)
    cluster.phase = ClusterPhase.OFFLINE.value
    cluster.save()